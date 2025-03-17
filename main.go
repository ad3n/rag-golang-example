package main

import (
	"log"

	"assistant/db"
	"assistant/embeddings"
	"assistant/handlers"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

type InsertRequest struct {
	Text string `json:"text"`
}

func main() {
	godotenv.Load()
	db.InitDB()

	app := fiber.New()

	app.Post("/store", func(c fiber.Ctx) error {
		var req InsertRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		if req.Text == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Text cannot be empty"})
		}

		err := embeddings.StoreEmbedding(db.DB, req.Text)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to store document"})
		}

		return c.JSON(fiber.Map{"message": "Document stored successfully"})
	})

	app.Post("/summaries", handlers.SummaryHandler())

	app.Get("/query", handlers.QueryHandler(db.DB))

	log.Fatal(app.Listen(":3000"))
}
