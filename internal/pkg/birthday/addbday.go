package birthday

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"discord-bot-go/internal/domain"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

func Add(args []string, s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB, command string) {
	var guildID string = m.GuildID

	var server domain.Server
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

		bday := domain.Birthday{User: args[1], Day: day, Month: month, ServerID: server.ID, Server: server}
		db.Create(&bday)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🎉 Aniversário de %s adicionado! 🎂", bday.User))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Formato correto: %s NOME DIA/MES ⚠️", command))
	}
}
