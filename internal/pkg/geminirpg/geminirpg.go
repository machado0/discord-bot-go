package geminirpg

import (
	"context"
	"discord-bot-go/internal/infra/localllm"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate, docContent string, botID string, chunkEmbeddings [][]float32, textChunks []string) {
	s.ChannelTyping(m.ChannelID)

	userPrompt := strings.Replace(m.Content, "<@"+botID+">", "", -1)
	userPrompt = strings.TrimSpace(userPrompt)

	log.Printf("Pergunta recebida de %s: %s", m.Author.Username, userPrompt)

	ctx := context.Background()

	questionEmbedding, err := localllm.GenerateEmbedding(ctx, userPrompt)
	if err != nil {
		log.Printf("Erro ao gerar embedding para a pergunta: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Desculpe, não consegui processar sua pergunta.")
		return
	}

	bestChunkIndex := FindBestChunk(questionEmbedding, chunkEmbeddings)
	if bestChunkIndex == -1 {
		log.Println("Não foi possível encontrar um chunk relevante.")
		s.ChannelMessageSend(m.ChannelID, "Desculpe, não encontrei nenhuma informação relevante sobre isso no documento.")
		return
	}
	contextChunk := textChunks[bestChunkIndex]
	log.Printf("Chunk mais relevante (índice %d) selecionado para o contexto.", bestChunkIndex)

	response, err := localllm.GenerateResponse(contextChunk, userPrompt)
	if err != nil {
		log.Printf("Erro ao gerar resposta da LLM: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Desculpe, ocorreu um erro ao tentar processar sua pergunta.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, response)
}
