package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
)

// GeminiResponse represents the structured JSON response from Gemini API
type GeminiResponse struct {
	Decision      string   `json:"decision"`
	Amount        *float64 `json:"amount"`
	Confidence    string   `json:"confidence"`
	Justification struct {
		PrimaryReasoning  string `json:"primary_reasoning"`
		SupportingClauses []struct {
			ClauseReference string `json:"clause_reference"`
			ClauseText      string `json:"clause_text"`
			Application     string `json:"application"`
		} `json:"supporting_clauses"`
		KeyFactors struct {
			EntityAnalysis struct {
				ExtractedEntities  []string `json:"extracted_entities"`
				MissingInformation []string `json:"missing_information"`
			} `json:"entity_analysis"`
			RuleApplication   string `json:"rule_application"`
			CalculationMethod string `json:"calculation_method"`
		} `json:"key_factors"`
	} `json:"justification"`
	Recommendations []string `json:"recommendations"`
	Flags           []string `json:"flags"`
}

// ValidateAndFormatResponse takes a JSON string, validates it, and returns formatted HTML
func ValidateAndFormatResponse(jsonStr string) (template.HTML, error) {
	// Try to parse the JSON into our struct
	var response GeminiResponse
	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return "", fmt.Errorf("invalid JSON format: %w", err)
	}

	// Build HTML response
	var html strings.Builder

	// Decision and Confidence section
	html.WriteString(`<div class="response-section">`)
	html.WriteString(fmt.Sprintf(`<h3 class="decision %s">Decision: %s</h3>`,
		strings.ToLower(response.Decision),
		response.Decision))
	html.WriteString(fmt.Sprintf(`<p class="confidence %s">Confidence: %s</p>`,
		strings.ToLower(response.Confidence),
		response.Confidence))

	// Amount if present
	if response.Amount != nil {
		html.WriteString(fmt.Sprintf(`<p class="amount">Amount: $%.2f</p>`, *response.Amount))
	}

	// Primary Reasoning
	html.WriteString(`<div class="reasoning-section">`)
	html.WriteString(fmt.Sprintf(`<p><strong>Primary Reasoning:</strong> %s</p>`,
		response.Justification.PrimaryReasoning))
	html.WriteString(`</div>`)

	// Supporting Clauses
	if len(response.Justification.SupportingClauses) > 0 {
		html.WriteString(`<div class="clauses-section">`)
		html.WriteString(`<h4>Supporting Clauses:</h4>`)
		html.WriteString(`<ul>`)
		for _, clause := range response.Justification.SupportingClauses {
			html.WriteString(`<li class="clause">`)
			html.WriteString(fmt.Sprintf(`<strong>%s</strong><br>`, clause.ClauseReference))
			html.WriteString(fmt.Sprintf(`<em>"%s"</em><br>`, clause.ClauseText))
			html.WriteString(fmt.Sprintf(`Application: %s`, clause.Application))
			html.WriteString(`</li>`)
		}
		html.WriteString(`</ul>`)
		html.WriteString(`</div>`)
	}

	// Key Factors
	html.WriteString(`<div class="factors-section">`)
	html.WriteString(`<h4>Key Factors:</h4>`)

	// Entity Analysis
	if len(response.Justification.KeyFactors.EntityAnalysis.ExtractedEntities) > 0 {
		html.WriteString(`<div class="entities">`)
		html.WriteString(`<strong>Extracted Entities:</strong><ul>`)
		for _, entity := range response.Justification.KeyFactors.EntityAnalysis.ExtractedEntities {
			html.WriteString(fmt.Sprintf(`<li>%s</li>`, entity))
		}
		html.WriteString(`</ul></div>`)
	}

	// Missing Information
	if len(response.Justification.KeyFactors.EntityAnalysis.MissingInformation) > 0 {
		html.WriteString(`<div class="missing-info">`)
		html.WriteString(`<strong>Missing Information:</strong><ul>`)
		for _, info := range response.Justification.KeyFactors.EntityAnalysis.MissingInformation {
			html.WriteString(fmt.Sprintf(`<li>%s</li>`, info))
		}
		html.WriteString(`</ul></div>`)
	}

	// Rule Application and Calculation Method
	html.WriteString(fmt.Sprintf(`<p><strong>Rule Application:</strong> %s</p>`,
		response.Justification.KeyFactors.RuleApplication))
	html.WriteString(fmt.Sprintf(`<p><strong>Calculation Method:</strong> %s</p>`,
		response.Justification.KeyFactors.CalculationMethod))
	html.WriteString(`</div>`)

	// Recommendations
	if len(response.Recommendations) > 0 {
		html.WriteString(`<div class="recommendations-section">`)
		html.WriteString(`<h4>Recommendations:</h4><ul>`)
		for _, rec := range response.Recommendations {
			html.WriteString(fmt.Sprintf(`<li>%s</li>`, rec))
		}
		html.WriteString(`</ul></div>`)
	}

	// Flags
	if len(response.Flags) > 0 {
		html.WriteString(`<div class="flags-section">`)
		html.WriteString(`<h4>Flags:</h4><ul>`)
		for _, flag := range response.Flags {
			html.WriteString(fmt.Sprintf(`<li>%s</li>`, flag))
		}
		html.WriteString(`</ul></div>`)
	}

	html.WriteString(`</div>`)

	return template.HTML(html.String()), nil
}
