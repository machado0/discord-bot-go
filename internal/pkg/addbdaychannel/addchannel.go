package addbdaychannel

import (
	"fmt"

	"discord-bot-go/internal/domain"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

func AddBirthdayChannel(s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB) {
	addServer(m.GuildID, m.ChannelID, db)
	s.ChannelMessageSend(m.ChannelID, "Canal adicionado com sucesso!")
}

func addServer(guildID string, birthdayChannel string, db *gorm.DB) domain.Server {
	var server domain.Server

	err := db.Where(domain.Server{Guild: guildID}).
		Assign(domain.Server{BirthdayChannel: birthdayChannel}).
		FirstOrCreate(&server).Error

	if err != nil {
		fmt.Println("Erro ao adicionar/atualizar servidor:", err)
	}

	return server
}