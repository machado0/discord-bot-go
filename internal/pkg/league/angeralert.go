package league

import (
	"discord-bot-go/internal/domain"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

func RegisterAnger(s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB) {
	angerAlert := domain.Anger{User: m.Author.Username}
	db.Create(&angerAlert)

	count := countAngerAlerts(db)
	lastAngerAlert := searchForLastAngerAlert(db)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Thumpy tiltou de novo! Estamos com %d registros de raiva e o último foi em: %s", count, lastAngerAlert.CreatedAt.Format("01/02/2006 às 15:04")))
}

func AngerAlertCounter(s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB) {
	count := countAngerAlerts(db)
	lastAngerAlert := searchForLastAngerAlert(db)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Estamos com %d registros de raiva e o último foi em: %s", count, lastAngerAlert.CreatedAt.Format("01/02/2006 às 15:04")))
}

func searchForLastAngerAlert(db *gorm.DB) domain.Anger {
	var lastAngerAlert domain.Anger
	db.Last(&lastAngerAlert)
	return lastAngerAlert
}

func countAngerAlerts(db *gorm.DB) int64 {
	var count int64
	db.Model(&domain.Anger{}).Count(&count)
	return count
}
