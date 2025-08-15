package verifybday

import (
	"fmt"
	"time"

	"discord-bot-go/internal/domain"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

func VerifyBirthday(s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB) {
	now := time.Now().Local()

	var birthdays []domain.Birthday
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