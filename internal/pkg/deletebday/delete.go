package deletebday

import (
	"fmt"

	"discord-bot-go/internal/domain"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

func Delete(args []string, s *discordgo.Session, m *discordgo.MessageCreate, db *gorm.DB, command string) {
	if len(args) > 1 &&
		len(args[1]) > 0 {
		var bday domain.Birthday
		db.Where("user = ?", args[1]).Delete(&bday)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("😥 Aniversário de %s removido! 😔", args[1]))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Formato correto: %s NOME DIA/MES ⚠️", command))
	}
}