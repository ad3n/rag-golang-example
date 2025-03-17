package handlers

import (
	"bytes"
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/ledongthuc/pdf"
)

func SummaryHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Failed to get file from form")
		}

		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to open uploaded file")
		}
		defer src.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, src)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to copy file to buffer")
		}

		text, err := extractTextFromPDF(bytes.NewReader(buf.Bytes()), file.Size)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to extract text from PDF")
		}

		prompt := "Buatkan rangkuman dari informasi berikut: \n" + text

		ollamaResp, err := generateResponse(prompt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate response"})
		}

		return c.JSON(fiber.Map{"answer": ollamaResp})
	}

}

func extractTextFromPDF(file io.ReaderAt, size int64) (string, error) {
	reader, err := pdf.NewReader(file, size)
	if err != nil {
		return "", err
	}

	totalPage := reader.NumPage()
	content := ""

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := reader.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		rows, _ := p.GetTextByRow()
		for _, row := range rows {
			for _, word := range row.Content {
				content += word.S
			}

			content += "\n"
		}
	}

	return content, nil
}
