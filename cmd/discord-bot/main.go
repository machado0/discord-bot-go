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
	_ "time/tzdata"

	"discord-bot-go/internal/database"
	"discord-bot-go/internal/domain"
	"discord-bot-go/internal/infra/gdocs"
	"discord-bot-go/internal/infra/riot"
	"discord-bot-go/internal/infra/localllm"
	"discord-bot-go/internal/pkg/birthday"
	"discord-bot-go/internal/pkg/geminirpg"
	"discord-bot-go/internal/pkg/league"
	"discord-bot-go/internal/pkg/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

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
	// carregando o arquivo .env
	if errorEnv := godotenv.Load(); errorEnv != nil {
		log.Println("*** Sem arquivo .env, utilizando variáveis locais! ***")
	}

	// tempo pra garantir que o DB subiu
	time.Sleep(5 * time.Second)

	database.Connect()
	db := database.DB

	botToken := os.Getenv("BOT_TOKEN")
	riotApiKey := os.Getenv("RIOT_API_KEY")
	botID := os.Getenv("BOT_ID")
	riotClient := riot.NewRiotClient(riotApiKey)

	docID := os.Getenv("GOOGLE_DOC_ID")
	if docID == "" {
		log.Fatal("Variável GOOGLE_DOC_ID não definida.")
	}

	sess, err := discordgo.New(fmt.Sprintf("Bot %s", botToken))
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		args := strings.Split(m.Content, " ")

		if m.Author.ID != botID || m.ChannelID == "1410982653972320267" {
			mentioned := false
			for _, user := range m.Mentions {
				if user.ID == botID {
					mentioned = true
					break
				}
			}

			if mentioned {
				log.Println("Carregando o conteúdo do Google Docs...")
				docContent, err := gdocs.ReadDocument(docID)
				if err != nil {
					log.Fatalf("Falha ao ler o Google Docs: %v", err)
				}

				rawChunks := strings.Split(docContent, "\n")
				var docChunks []string
				for _, chunk := range rawChunks {
					trimmedChunk := strings.TrimSpace(chunk)
					// Ignora chunks muito pequenos que não são úteis
					if len(trimmedChunk) > 10 {
						docChunks = append(docChunks, trimmedChunk)
					}
				}

				if len(docChunks) == 0 {
					log.Fatal("Nenhum conteúdo útil (chunks) foi encontrado no documento.")
				}
				log.Printf("Documento dividido em %d chunks.", len(docChunks))

				log.Println("Gerando embeddings para os chunks do documento... (Isso pode demorar um pouco)")
				startTime := time.Now()
				var chunkEmbeddings [][]float32
				ctx := context.Background()

				for i, chunk := range docChunks {
					embedding, err := localllm.GenerateEmbedding(ctx, chunk)
					if err != nil {
						log.Fatalf("Falha ao gerar embedding para o chunk %d: %v", i, err)
					}
					chunkEmbeddings = append(chunkEmbeddings, embedding)
					log.Printf("Embedding gerado para o chunk %d/%d", i+1, len(docChunks))
				}
				duration := time.Since(startTime)
				log.Printf("Todos os %d embeddings foram gerados com sucesso em %s.", len(docChunks), duration)
				geminirpg.MessageCreate(s, m, docContent, botID, chunkEmbeddings, docChunks)
			}
		} else {
			switch args[0] {
			case command_add:
				birthday.Add(args, s, m, db, command_add)
			case command_delete:
				birthday.Delete(args, s, m, db, command_add)
			case command_next_bday:
				birthday.NextBirthday(s, m, db, command_next_bday)
			case command_list:
				birthday.List(s, m, db, command_list)
			case command_add_channel:
				birthday.AddBirthdayChannel(s, m, db)
			case command_verify_bdays:
				birthday.VerifyBirthday(s, m, db)
			case command_commands:
				utils.ListCommands(s, m)
			case command_verify_thumpy_soloduo:
				league.VerifyThumpyCommand(s, m, riotClient)
			case command_anger_alert:
				league.RegisterAnger(s, m, db)
			case command_anger_counter:
				league.AngerAlertCounter(s, m, db)
			}
		}

	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("Bot online!")

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		dailyBirthdayCheck(sess, db)
	}()

	go func() {
		// thumpyCheckEveryTwoMinutes(ctx, sess, db, riotClient)
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
		brtLocation, _ := time.LoadLocation("America/Sao_Paulo")
		now := time.Now().In(brtLocation)
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), targetHour, targetMinute, targetSecond, 0, brtLocation)

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

			_, _, matchId := league.GetThumpyLastSoloDuo(client)

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

			_, matchResponse, currentMatchId := league.GetThumpyLastSoloDuo(client)

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
				league.SendDiscordMessageAfterMatch(client, s, matchResponse, thumpyPuuid)
			} else {
				fmt.Println("Nenhuma nova partida encontrada")
			}
		}
	}
}
