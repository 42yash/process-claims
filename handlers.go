package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling home request from %s", r.RemoteAddr)
	tmpl := template.Must(template.ParseFiles("templates/home.html"))

	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleSubmit processes the form submission with query, system prompt, and PDF
func handleSubmit(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling submit request from %s", r.RemoteAddr)

	// Parse the multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Extract form values
	query := r.FormValue("query")
	if query == "" {
		log.Print("Empty query submitted")
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
	log.Print("Reading uploaded PDF file")
	pdfBytes, err := readFileBytes(file)
	if err != nil {
		log.Printf("Error reading PDF file: %v", err)
		http.Error(w, "Error reading PDF file", http.StatusInternalServerError)
		return
	}

	// Read system prompt from file
	log.Print("Reading system prompt")
	systemPromptBytes, err := os.ReadFile("system_prompt.txt")
	if err != nil {
		log.Printf("Error reading system prompt: %v", err)
		http.Error(w, "Error reading system prompt file", http.StatusInternalServerError)
		return
	}
	systemPrompt := string(systemPromptBytes)

	// Process with Gemini API
	log.Print("Processing with Gemini API")
	response, err := processWithGemini(r.Context(), query, systemPrompt, pdfBytes)
	if err != nil {
		log.Printf("Error in Gemini API processing: %v", err)
		http.Error(w, fmt.Sprintf("Error processing with Gemini: %v", err), http.StatusInternalServerError)
		return
	}
	log.Print("Successfully processed with Gemini API")

	// Format and validate the response
	formattedResponse, err := ValidateAndFormatResponse(response)
	if err != nil {
		log.Printf("Error formatting response: %v", err)
		http.Error(w, fmt.Sprintf("Error formatting response: %v", err), http.StatusInternalServerError)
		return
	}

	responseStr := string(formattedResponse)
	if responseStr == "" {
		log.Print("Warning: Empty formatted response")
		http.Error(w, "Empty response from processing", http.StatusInternalServerError)
		return
	}

	// Log the first 100 characters of the response
	if len(responseStr) > 100 {
		log.Printf("Sending formatted response (first 100 chars): %s...", responseStr[:100])
	} else {
		log.Printf("Sending formatted response: %s", responseStr)
	}

	// Set content type and send response
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div class="response-content">%s</div>`, formattedResponse)
}
