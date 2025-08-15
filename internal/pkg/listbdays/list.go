package listbirthdays

import (
	"fmt"
	"log"

	"discord-bot-go/internal/domain"
	"discord-bot-go/internal/pkg/utils"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

func List(s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB, command string) {
	var server domain.Server = utils.FindServer(m.GuildID, db)

	var bdays []domain.Birthday
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