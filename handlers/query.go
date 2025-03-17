package handlers

import (
	"assistant/embeddings"
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/ollama/ollama/api"
	"github.com/pgvector/pgvector-go"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func QueryHandler(db *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		query := c.Query("q")
		if query == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Query parameter is required"})
		}

		embedding, err := embeddings.GetEmbedding(query)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate embedding"})
		}

		rows, err := db.Query(`SELECT content FROM documents ORDER BY embedding <-> $1 LIMIT 3`, pgvector.NewVector(embedding))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to retrieve documents"})
		}
		defer rows.Close()

		var contextText string
		for rows.Next() {
			var content string
			rows.Scan(&content)
			contextText += content + "\n"
		}

		prompt := "Gunakan informasi berikut: \n" + contextText + "\n Untuk menjawab pertanyaan: " + query
		ollamaResp, err := generateResponse(prompt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate response"})
		}

		return c.JSON(fiber.Map{"answer": ollamaResp})
	}
}

func generateResponse(prompt string) (string, error) {
	client := api.NewClient(&url.URL{
		Scheme: "http",
		Host:   "localhost:11434",
	}, http.DefaultClient)

	response := ""
	stream := false
	err := client.Chat(context.Background(), &api.ChatRequest{
		Model:  "llama3.2",
		Stream: &stream,
		Messages: []api.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}, func(cr api.ChatResponse) error {
		response = strings.TrimSpace(cr.Message.Content)

		return nil
	})

	return response, err
}
