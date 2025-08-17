package main

import (
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
	"discord-bot-go/internal/pkg/deletebday"
	"discord-bot-go/internal/pkg/listbdays"
	"discord-bot-go/internal/pkg/nextbday"
	"discord-bot-go/internal/pkg/verifybday"
	"discord-bot-go/internal/pkg/utils"
	"discord-bot-go/internal/pkg/verifythumpysoloqueue"

	"github.com/bwmarrin/discordgo"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
)

// TODO: fazer o cadastrar canal funcionar

var db, err = gorm.Open(sqlite.Open(os.Getenv("DB_NAME")), &gorm.Config{})

var command_add string = "!adicionar"
var command_delete string = "!remover"
var command_next_bday string = "!proximo"
var command_list string = "!listar"
var command_verify_bdays string = "!verificar"
var command_add_channel string = "!addcanal"
var command_commands string = "!comandos"
var command_verificar_solo_duo_luca string = "!soloduo"

func main() {
	db.AutoMigrate(&domain.Birthday{})

	errorEnv := godotenv.Load()
		if errorEnv != nil {
			log.Fatalf("Erro carregando arquivo .env: %v", errorEnv)
		}

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
		case command_verificar_solo_duo_luca:
			verifythumpysoloqueue.VerifyThumpySoloQueueCommand(s, m, riotClient)
		}

	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer sess.Close()
	fmt.Println("bot online!")

	dailyBirthdayCheck(sess)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

func dailyBirthdayCheck(s *discordgo.Session) {
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

		fmt.Println("rodou bot de aniversario:", time.Now())

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