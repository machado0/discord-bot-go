package studycounter

import (
	"discord-bot-go/internal/domain"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

func Createstudysession(s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB, date *time.Time) {
	var server domain.Server
	db.Where("guild = ?", m.GuildID).Find(&server)

	session := domain.StudySession{Server: server, ServerID: server.ID, Date: date}
	db.Create(&session)

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sessão de estudos criada para %v", date))
}

func Registerstudychannel(s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB) {
	addServer(m.GuildID, m.ChannelID, db)
	s.ChannelMessageSend(m.ChannelID, "Canal adicionado com sucesso!")
}

func addServer(guildID string, studychannel string, db *gorm.DB) domain.Server {
	var server domain.Server

	err := db.Where(domain.Server{Guild: guildID}).
		Assign(domain.Server{StudyChannel: studychannel}).
		FirstOrCreate(&server).Error

	if err != nil {
		log.Println("Erro ao adicionar/atualizar servidor para adicionar canal de estudos:", err)
	}

	return server
}
