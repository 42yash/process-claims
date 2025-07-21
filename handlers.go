package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"google.golang.org/genai"
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/home.html"))

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleSubmit processes the form submission with query, system prompt, and PDF
func handleSubmit(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Extract form values
	query := r.FormValue("query")
	if query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, _, err := r.FormFile("document_file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read PDF content
	pdfBytes, err := readFileBytes(file)
	if err != nil {
		http.Error(w, "Error reading PDF file", http.StatusInternalServerError)
		return
	}

	// Read system prompt from file
	systemPromptBytes, err := os.ReadFile("system_prompt.txt")
	if err != nil {
		http.Error(w, "Error reading system prompt file", http.StatusInternalServerError)
		return
	}
	systemPrompt := string(systemPromptBytes)

	// Process with Gemini API
	response, err := processWithGemini(r.Context(), query, systemPrompt, pdfBytes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing with Gemini: %v", err), http.StatusInternalServerError)
		return
	}

	// Set content type for HTML response
	w.Header().Set("Content-Type", "text/html")

	// Return formatted response
	fmt.Fprintf(w, `
		<div class="response-content">
			<h3>Analysis Results</h3>
			<div class="analysis-result">
				<pre>%s</pre>
			</div>
		</div>
	`, response)
}

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
