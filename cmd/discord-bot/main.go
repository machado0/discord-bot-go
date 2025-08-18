package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"discord-bot-go/internal/domain"
	"discord-bot-go/internal/infra/riot"
	"discord-bot-go/internal/pkg/addbday"
	"discord-bot-go/internal/pkg/addbdaychannel"
	"discord-bot-go/internal/pkg/angeralert"
	"discord-bot-go/internal/pkg/deletebday"
	listbirthdays "discord-bot-go/internal/pkg/listbdays"
	"discord-bot-go/internal/pkg/nextbday"
	"discord-bot-go/internal/pkg/utils"
	"discord-bot-go/internal/pkg/verifybday"
	"discord-bot-go/internal/pkg/verifythumpy"

	"github.com/bwmarrin/discordgo"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// TODO: fazer o cadastrar canal funcionar

var command_add string = "!adicionar"
var command_delete string = "!remover"
var command_next_bday string = "!proximo"
var command_list string = "!listar"
var command_verify_bdays string = "!verificar"
var command_add_channel string = "!addcanal"
var command_commands string = "!comandos"
var command_verify_thumpy_soloduo string = "!soloduo"
var command_anger_alert string = "!tiltou"
var command_anger_counter string = "!rages"

func main() {
	errorEnv := godotenv.Load()
	if errorEnv != nil {
		log.Fatalf("Erro carregando arquivo .env: %v", errorEnv)
	}

	var db, err = gorm.Open(sqlite.Open(os.Getenv("DB_NAME")), &gorm.Config{})
	db.AutoMigrate(&domain.Birthday{}, &domain.LastMatch{}, &domain.Server{}, &domain.Anger{})

	botToken := os.Getenv("BOT_TOKEN")
	riotApiKey := os.Getenv("RIOT_API_KEY")
	riotClient := riot.NewRiotClient(riotApiKey)

	sess, err := discordgo.New(fmt.Sprintf("Bot %s", botToken))
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		args := strings.Split(m.Content, " ")
		switch args[0] {
		case command_add:
			addbday.Add(args, s, m, db, command_add)
		case command_delete:
			deletebday.Delete(args, s, m, db, command_add)
		case command_next_bday:
			nextbday.NextBirthday(s, m, db, command_next_bday)
		case command_list:
			listbirthdays.List(s, m, db, command_list)
		case command_add_channel:
			addbdaychannel.AddBirthdayChannel(s, m, db)
		case command_verify_bdays:
			verifybday.VerifyBirthday(s, m, db)
		case command_commands:
			utils.ListCommands(s, m)
		case command_verify_thumpy_soloduo:
			verifythumpy.VerifyThumpyCommand(s, m, riotClient)
		case command_anger_alert:
			angeralert.RegisterAnger(s, m, db)
		case command_anger_counter:
			angeralert.AngerAlertCounter(s, m, db)
		}
	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("Bot online!")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		dailyBirthdayCheck(sess, db)
	}()

	go func() {
		thumpyCheckEveryTwoMinutes(ctx, sess, db, riotClient)
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	fmt.Println("Pressione Ctrl+C para parar o bot...")
	<-sc

	fmt.Println("Parando o bot...")
	cancel()

	time.Sleep(1 * time.Second)
	fmt.Println("Bot parado!")
}

func dailyBirthdayCheck(s *discordgo.Session, db *gorm.DB) {
	targetHour := 9
	targetMinute := 0
	targetSecond := 0

	for {
		now := time.Now().Local()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), targetHour, targetMinute, targetSecond, 0, time.Local)

		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		durationUntilNextRun := nextRun.Sub(now)
		time.Sleep(durationUntilNextRun)

		fmt.Println("rodou rotina de aniversario:", time.Now())

		var birthdays []domain.Birthday
		result := db.Where("day = ? AND month = ?", now.Day(), int(now.Month())).Preload("Server").Find(&birthdays)

		if result.Error != nil {
			fmt.Println("Erro ao buscar aniversariantes:", result.Error)
			continue
		}

		for _, bday := range birthdays {
			if len(bday.Server.BirthdayChannel) > 1 {
				s.ChannelMessageSend(bday.Server.BirthdayChannel, fmt.Sprintf("@everyone 🎉 Hoje é aniversário de %s! 🎂", bday.User))
			}
		}
	}
}

func thumpyCheckEveryTwoMinutes(ctx context.Context, s *discordgo.Session, db *gorm.DB, client *riot.RiotClient) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	fmt.Println("Iniciando verificação thumpy")
	thumpyPuuid := os.Getenv("THUMPY")

	if thumpyPuuid == "" {
		fmt.Println("ERRO: THUMPY puuid não configurado no .env")
		return
	}

	var lastMatch domain.LastMatch
	result := db.Where("puuid = ?", thumpyPuuid).First(&lastMatch)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			fmt.Println("Nenhuma partida anterior encontrada para Thumpy")

			_, _, matchId := verifythumpy.GetThumpyLastSoloDuo(client)

			if matchId != "" {
				obj := domain.LastMatch{
					MatchId: matchId,
					Puuid:   thumpyPuuid,
				}

				if err := db.Create(&obj).Error; err != nil {
					fmt.Printf("Erro ao criar registro inicial: %v\n", err)
					return
				}

				lastMatch = obj
				fmt.Println("Primeira partida registrada:", matchId)
			} else {
				fmt.Println("Não foi possível obter partida inicial")
				return
			}
		} else {
			fmt.Printf("Erro ao buscar última partida do Thumpy: %v\n", result.Error)
			return
		}
	} else {
		fmt.Println("Última partida conhecida:", lastMatch.MatchId)
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Parando verificação Thumpy")
			return
		case <-ticker.C:
			fmt.Println("Verificando partidas do Thumpy:", time.Now().Format("15:04:05"))

			_, matchResponse, currentMatchId := verifythumpy.GetThumpyLastSoloDuo(client)

			if currentMatchId == "" {
				fmt.Println("Erro ao obter partida atual da API")
				continue
			}

			if currentMatchId != lastMatch.MatchId {
				fmt.Printf("Nova partida detectada! Anterior: %s, Nova: %s\n", lastMatch.MatchId, currentMatchId)

				if err := db.Model(&lastMatch).Update("match_id", currentMatchId).Error; err != nil {
					fmt.Printf("Erro ao atualizar partida no banco: %v\n", err)
					continue
				}

				lastMatch.MatchId = currentMatchId
				fmt.Printf("Partida atualizada com sucesso: %s\n", currentMatchId)
				verifythumpy.SendDiscordMessageAfterMatch(client, s, matchResponse, thumpyPuuid)
			} else {
				fmt.Println("Nenhuma nova partida encontrada")
			}
		}
	}
}
