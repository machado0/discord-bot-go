package gdocs

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

func ReadDocument(docID string) (string, error) {
	ctx := context.Background()

	privateKey := os.Getenv("GCP_SERVICE_ACCOUNT_PRIVATE_KEY")
	clientEmail := os.Getenv("GCP_SERVICE_ACCOUNT_CLIENT_EMAIL")
	projectID := os.Getenv("GEMINI_PROJECT_ID")

	if privateKey == "" || clientEmail == "" || projectID == "" {
		return "", fmt.Errorf("as variáveis de ambiente para a conta de serviço do Google Cloud (GCP_SERVICE_ACCOUNT_PRIVATE_KEY, GCP_SERVICE_ACCOUNT_CLIENT_EMAIL, GCP_SERVICE_ACCOUNT_PROJECT_ID) devem ser definidas")
	}

	privateKey = strings.ReplaceAll(privateKey, "\\n", "\n")

	credentialsJSON := fmt.Sprintf(`{
		"type": "service_account",
		"project_id": "%s",
		"private_key": %q,
		"client_email": "%s",
		"token_uri": "https://oauth2.googleapis.com/token"
	}`, projectID, privateKey, clientEmail)

	creds, err := google.CredentialsFromJSON(ctx, []byte(credentialsJSON), docs.DocumentsReadonlyScope)
	if err != nil {
		return "", fmt.Errorf("falha ao criar credenciais a partir das variáveis de ambiente: %w", err)
	}

	docsService, err := docs.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return "", fmt.Errorf("falha ao criar serviço do Docs: %w", err)
	}

	doc, err := docsService.Documents.Get(docID).Do()
	if err != nil {
		return "", fmt.Errorf("falha ao obter o documento (ID: %s): %w. Verifique se o ID está correto e se o documento foi compartilhado com o e-mail da service account", docID, err)
	}

	var contentBuilder strings.Builder
	for _, elem := range doc.Body.Content {
		if elem.Paragraph != nil {
			for _, pe := range elem.Paragraph.Elements {
				if pe.TextRun != nil {
					contentBuilder.WriteString(pe.TextRun.Content)
				}
			}
		}
	}

	return contentBuilder.String(), nil
}

