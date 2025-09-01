package gemini

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func generateClient(ctx context.Context) (*genai.Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("variável de ambiente GEMINI_API_KEY não definida")
	}
	return genai.NewClient(ctx, option.WithAPIKey(apiKey))
}

func GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	client, err := generateClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	em := client.EmbeddingModel("embedding-001")
	res, err := em.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar embedding: %w", err)
	}

	if res == nil || res.Embedding == nil || len(res.Embedding.Values) == 0 {
		return nil, fmt.Errorf("resposta de embedding vazia ou inválida")
	}

	return res.Embedding.Values, nil
}

func GenerateResponse(contextChunk, userPrompt string) (string, error) {
	ctx := context.Background()
	client, err := generateClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-pro")

	systemPrompt := fmt.Sprintf(`
		Você é um narrador prestativo. Sua tarefa é responder à pergunta do usuário.
		Use o trecho de contexto fornecido abaixo como sua principal fonte de informação.
		Responda de forma concisa e baseie-se estritamente no contexto.
		Se a resposta não estiver no contexto, diga "O Ale precisa trabalhar mais..."

		--- CONTEXTO ---
		%s
		--- FIM DO CONTEXTO ---
	`, contextChunk)

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(systemPrompt),
		},
	}

	resp, err := model.GenerateContent(ctx, genai.Text(userPrompt))
	if err != nil {
		return "", fmt.Errorf("erro ao gerar conteúdo do Gemini: %w", err)
	}

	var responseBuilder strings.Builder
	if resp != nil && len(resp.Candidates) > 0 {
		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					if txt, ok := part.(genai.Text); ok {
						responseBuilder.WriteString(string(txt))
					}
				}
			}
		}
	} else {
		return "", fmt.Errorf("resposta do Gemini vazia ou em formato inesperado")
	}

	responseText := responseBuilder.String()
	if responseText == "" {
		log.Println("A resposta do Gemini não continha texto extraível.")
		return "Não consegui gerar uma resposta.", nil
	}

	return responseText, nil
}

