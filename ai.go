package main

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"google.golang.org/genai"
)

// processWithGemini handles the actual API call to Gemini
func processWithGemini(ctx context.Context, query, systemPrompt string, pdfBytes []byte) (string, error) {
	// Create Gemini client using the new Gen AI SDK
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create genai client: %w", err)
	}

	// Create multipart content with system prompt, user query, and PDF
	parts := []*genai.Part{
		{Text: fmt.Sprintf("System: %s", systemPrompt)},
		{Text: fmt.Sprintf("User Query: %s", query)},
		{InlineData: &genai.Blob{
			Data:     pdfBytes,
			MIMEType: "application/pdf",
		}},
	}

	// Create content payload
	content := []*genai.Content{{Parts: parts}}

	// Call Gemini 2.5 Flash model (latest as of July 2025)
	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", content, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return resp.Text(), nil
}

// readFileBytes reads all bytes from a multipart file
func readFileBytes(file multipart.File) ([]byte, error) {
	return io.ReadAll(file)
}
