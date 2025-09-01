package localllm

import (
	"bytes"
	"context"
	"discord-bot-go/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	apiURL := os.Getenv("LOCAL_LLM_API_URL")
	model := os.Getenv("LOCAL_LLM_EMBEDDING_MODEL")
	apiKey := os.Getenv("LOCAL_LLM_API_KEY")

	if apiURL == "" || model == "" {
		return nil, fmt.Errorf("as variáveis LOCAL_LLM_API_URL e LOCAL_LLM_EMBEDDING_MODEL devem ser definidas")
	}

	endpoint := fmt.Sprintf("%s/v1/embeddings", strings.TrimSuffix(apiURL, "/"))

	reqPayload := domain.OpenAIEmbeddingRequest{
		Model: model,
		Input: text,
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("erro ao codificar payload de embedding: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de embedding: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição de embedding para LLM local: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM local retornou erro de embedding %d: %s", resp.StatusCode, string(body))
	}

	var respPayload domain.OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta de embedding: %w", err)
	}

	if len(respPayload.Data) == 0 || len(respPayload.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("resposta de embedding do LLM local está vazia")
	}

	return respPayload.Data[0].Embedding, nil
}

func GenerateResponse(contextChunk, userPrompt string) (string, error) {
	apiURL := os.Getenv("LOCAL_LLM_API_URL")
	model := os.Getenv("LOCAL_LLM_CHAT_MODEL")
	apiKey := os.Getenv("LOCAL_LLM_API_KEY")

	if apiURL == "" || model == "" {
		return "", fmt.Errorf("as variáveis LOCAL_LLM_API_URL e LOCAL_LLM_CHAT_MODEL devem ser definidas")
	}

	endpoint := fmt.Sprintf("%s/v1/chat/completions", strings.TrimSuffix(apiURL, "/"))

	systemPrompt := fmt.Sprintf(`
		Você é um narrador prestativo. Sua tarefa é responder à pergunta do usuário.
		Use o trecho de contexto fornecido abaixo como sua principal fonte de informação.
		Responda de forma concisa e baseie-se estritamente no contexto.
		Se a resposta não estiver no contexto, diga "A informação para responder a isso não foi encontrada no trecho fornecido."

		--- CONTEXTO ---
		%s
		--- FIM DO CONTEXTO ---
	`, contextChunk)

	messages := []domain.OpenAIMessage{
		{Role: "user", Content: systemPrompt}, //switched to user cuz im using mistral, but use "system" if needed
		{Role: "user", Content: userPrompt},
	}

	reqPayload := domain.OpenAICompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("erro ao codificar payload de chat: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição de chat: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao enviar requisição de chat para LLM local: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM local retornou erro de chat %d: %s", resp.StatusCode, string(body))
	}

	var respPayload domain.OpenAICompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta de chat: %w", err)
	}

	if len(respPayload.Choices) == 0 {
		log.Println("A resposta do LLM local não continha texto extraível.")
		return "Não consegui gerar uma resposta.", nil
	}

	return respPayload.Choices[0].Message.Content, nil
}
