package embeddings

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/ollama/ollama/api"
	"github.com/pgvector/pgvector-go"
)

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func GetEmbedding(text string) ([]float32, error) {
	client := api.NewClient(&url.URL{
		Scheme: "http",
		Host:   os.Getenv("OLLAMA_HOST"),
	}, http.DefaultClient)

	response, err := client.Embeddings(context.Background(), &api.EmbeddingRequest{
		Model:  "llama3.2",
		Prompt: text,
	})

	if err != nil {
		return nil, err
	}

	e := make([]float32, len(response.Embedding))
	for i, f := range response.Embedding {
		e[i] = float32(f)
	}

	return e, nil
}

func StoreEmbedding(db *sql.DB, text string) error {
	embedding, err := GetEmbedding(text)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO documents (content, embedding) VALUES ($1, $2)",
		text, pgvector.NewVector(embedding),
	)
	if err != nil {
		log.Println("Error inserting document:", err)
		return err
	}

	log.Println("Document stored successfully!")
	return nil
}
