package nextbday

import (
	"fmt"
	"log"
	"sort"
	"time"

	"discord-bot-go/internal/domain"
	"discord-bot-go/internal/pkg/utils"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

func NextBirthday(s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB, command string) {
	var server domain.Server = utils.FindServer(m.GuildID, db)

	var bdays []domain.Birthday
	result := db.Where("server_id = ?", server.ID).Find(&bdays)

	if result.RowsAffected == 0 {
		log.Println("Nenhum aniversário encontrado para este servidor.")
		return
	}

	var bdaysAfterToday []domain.Birthday
	var today time.Time = time.Now()

	for _, bday := range bdays {
		if today.Local().Before(time.Date(today.Year(), time.Month(bday.Month), bday.Day, 0, 0, 0, 0, time.Local)) {
			bdaysAfterToday = append(bdaysAfterToday, bday)
		}
	}

	sort.Slice(bdaysAfterToday, func(i, j int) bool {
		date1 := time.Date(utils.CalculateYear(today, bdaysAfterToday[i].Month), time.Month(bdaysAfterToday[i].Month), bdaysAfterToday[i].Day, 0, 0, 0, 0, time.Local)
		date2 := time.Date(utils.CalculateYear(today, bdaysAfterToday[j].Month), time.Month(bdaysAfterToday[j].Month), bdaysAfterToday[j].Day, 0, 0, 0, 0, time.Local)
		return date1.Before(date2)
	})

	var nextBdays []domain.Birthday

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