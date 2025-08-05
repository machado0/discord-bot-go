package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	// "github.com/joho/godotenv"
)

// TODO: fazer o cadastrar canal funcionar

type Birthday struct {
	gorm.Model
	User     string
	Day      int
	Month    int
	ServerID uint
	Server   Server
}

type Server struct {
	gorm.Model
	Guild           string
	BirthdayChannel string
}

var db, err = gorm.Open(sqlite.Open(os.Getenv("DB_NAME")), &gorm.Config{})

var command_add string = "!adicionar"
var command_delete string = "!remover"
var command_next_bday string = "!proximo"
var command_list string = "!listar"
var command_verify_bdays string = "!verificar"
var command_add_channel string = "!addcanal"
var command_commands string = "!comandos"

func main() {
	db.AutoMigrate(&Birthday{})

	// errorEnv := godotenv.Load()
	// 	if errorEnv != nil {
	// 		log.Fatalf("Erro carregando arquivo .env: %v", errorEnv)
	// 	}

	botToken := os.Getenv("BOT_TOKEN")

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
			add(args, s, m)
		case command_delete:
			delete(args, s, m)
		case command_next_bday:
			nextBirthday(s, m)
		case command_list:
			list(s, m)
		case command_add_channel:
			addBirthdayChannel(s, m)
		case command_verify_bdays:
			verifyBirthday(s, m)
		case command_commands:
			listCommands(s, m)
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

func add(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	var guildID string = m.GuildID

	var server Server
	db.Where("guild = ?", guildID).Find(&server)

	if len(args) > 2 &&
		len(args[1]) > 0 &&
		len(args[2]) > 0 {
		date := strings.Split(args[2], "/")

		day, err := strconv.Atoi(date[0])
		if err != nil {
			log.Fatal(err)
		}

		month, err := strconv.Atoi(date[1])
		if err != nil {
			log.Fatal(err)
		}

		bday := Birthday{User: args[1], Day: day, Month: month, ServerID: server.ID, Server: server}
		db.Create(&bday)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🎉 Aniversário de %s adicionado! 🎂", bday.User))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Formato correto: %s NOME DIA/MES ⚠️", command_add))
	}
}

func delete(args []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(args) > 1 &&
		len(args[1]) > 0 {
		var bday Birthday
		db.Where("user = ?", args[1]).Delete(&bday)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("😥 Aniversário de %s removido! 😔", args[1]))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Formato correto: %s NOME DIA/MES ⚠️", command_delete))
	}
}

func list(s *discordgo.Session, m *discordgo.MessageCreate) {
	var server Server = findServer(m.GuildID)

	var bdays []Birthday
	result := db.Where("server_id = ?", server.ID).Find(&bdays)

	if result.RowsAffected == 0 {
		log.Println("Nenhum aniversário encontrado para este servidor.")
		return
	}

	var names string = "=================================="

	for _, bday := range bdays {
		names += fmt.Sprintf("\n%s - %d/%d", bday.User, bday.Day, bday.Month)
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("📋 Lista de Aniversários do Servidor 📋\n %s", names))
}

func nextBirthday(s *discordgo.Session, m *discordgo.MessageCreate) {
	var server Server = findServer(m.GuildID)

	var bdays []Birthday
	result := db.Where("server_id = ?", server.ID).Find(&bdays)

	if result.RowsAffected == 0 {
		log.Println("Nenhum aniversário encontrado para este servidor.")
		return
	}

	var bdaysAfterToday []Birthday
	var today time.Time = time.Now()

	for _, bday := range bdays {
		if today.Local().Before(time.Date(today.Year(), time.Month(bday.Month), bday.Day, 0, 0, 0, 0, time.Local)) {
			bdaysAfterToday = append(bdaysAfterToday, bday)
		}
	}

	sort.Slice(bdaysAfterToday, func(i, j int) bool {
		date1 := time.Date(calculateYear(today, bdaysAfterToday[i].Month), time.Month(bdaysAfterToday[i].Month), bdaysAfterToday[i].Day, 0, 0, 0, 0, time.Local)
		date2 := time.Date(calculateYear(today, bdaysAfterToday[j].Month), time.Month(bdaysAfterToday[j].Month), bdaysAfterToday[j].Day, 0, 0, 0, 0, time.Local)
		return date1.Before(date2)
	})

	var nextBdays []Birthday

	for _, bday := range bdaysAfterToday {
		if bday.Day == bdaysAfterToday[0].Day {
			nextBdays = append(nextBdays, bday)
		}
	}

	var names string

	if len(nextBdays) > 1 {
		for _, bday := range bdays {
			names += fmt.Sprintf("%s,", bday.User)
		}
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("📋 O próximo aniversário é de %s em %d/%d 📋 ", names, nextBdays[0].Day, nextBdays[0].Month))
}

func addBirthdayChannel(s *discordgo.Session, m *discordgo.MessageCreate) {
	addServer(m.GuildID, m.ChannelID)
	s.ChannelMessageSend(m.ChannelID, "Canal adicionado com sucesso!")
}

func verifyBirthday(s *discordgo.Session, m *discordgo.MessageCreate) {
	now := time.Now().Local()

	var birthdays []Birthday
	result := db.Joins("JOIN servers ON servers.id = birthdays.server_id").Where("birthdays.day = ? AND birthdays.month = ? AND servers.guild = ?", now.Day(), int(now.Month()), m.GuildID).Preload("Server").Find(&birthdays)

	if result.Error != nil {
		fmt.Println("Erro ao buscar aniversariantes:", result.Error)
	}

	if len(birthdays) == 0 {
		s.ChannelMessageSend(m.ChannelID, "😞 Sem aniversários neste dia 😢")
	}

	for _, bday := range birthdays {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("@everyone 🎉 Hoje é aniversário de %s! 🎂", bday.User))
	}
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

		fmt.Println("rodou bot", time.Now())

		var birthdays []Birthday
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

func listCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	var sb strings.Builder

	sb.WriteString("=== Comandos do Bot! ===\n")
	sb.WriteString("!adicionar __***NOME DIA/MES***__ - Adiciona um aniversário\n")
	sb.WriteString("!remover __***NOME***__ - Remove um aniversário\n")
	sb.WriteString("!proximo - Lista o próximo aniversário\n")
	sb.WriteString("!listar - Lista todos os aniversários\n")
	sb.WriteString("!verificar - Força a verificação de aniversário\n")
	sb.WriteString("!addcanal - Configura o canal atual como canal do bot\n")
	sb.WriteString("!comandos - Lista os comandos do bot\n")

	s.ChannelMessageSend(m.ChannelID, sb.String())
}

func addServer(guildID string, birthdayChannel string) Server {
	var server Server

	err := db.Where(Server{Guild: guildID}).
		Assign(Server{BirthdayChannel: birthdayChannel}).
		FirstOrCreate(&server).Error

	if err != nil {
		fmt.Println("Erro ao adicionar/atualizar servidor:", err)
	}

	return server
}

func findServer(guildID string) Server {
	var server Server
	db.Where("guild = ?", guildID).Find(&server)

	return server
}

func calculateYear(today time.Time, todaysMonth int) int {
	if todaysMonth == 12 {
		return today.Year() + 1
	} else {
		return today.Year()
	}
}
