package utils

import (
	"strings"
	"time"

	"discord-bot-go/internal/domain"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

func FindServer(guildID string, db *gorm.DB) domain.Server {
	var server domain.Server
	db.Where("guild = ?", guildID).Find(&server)

	return server
}

func CalculateYear(today time.Time, todaysMonth int) int {
	if todaysMonth == 12 {
		return today.Year() + 1
	} else {
		return today.Year()
	}
}

func ListCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	var sb strings.Builder

	sb.WriteString("=== Comandos do Bot! ===\n")
	sb.WriteString("!adicionar __***NOME***__ __***DIA/MES***__ - Adiciona um aniversário\n")
	sb.WriteString("!remover __***NOME***__ - Remove um aniversário\n")
	sb.WriteString("!proximo - Lista o próximo aniversário\n")
	sb.WriteString("!listar - Lista todos os aniversários\n")
	sb.WriteString("!verificar - Força a verificação de aniversário\n")
	sb.WriteString("!addcanal - Configura o canal atual como canal do bot\n")
	sb.WriteString("!comandos - Lista os comandos do bot\n")

	s.ChannelMessageSend(m.ChannelID, sb.String())
}