package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type AnalyzeRequest struct {
	Text         string   `json:"text"`
	Requirements []string `json:"requirements,omitempty"`
}

type RequirementResult struct {
	Requirement   string   `json:"requirement"`
	Matched       bool     `json:"matched"`
	Score         float64  `json:"score"`
	KeywordsFound []string `json:"keywords_found"`
}

type SentenceLength struct {
	Avg float64 `json:"avg"`
	Min int     `json:"min"`
	Max int     `json:"max"`
	Std float64 `json:"std"`
}

type Paragraphs struct {
	ParagraphCount  int  `json:"paragraph_count"`
	HasIntroduction bool `json:"has_introduction"`
	HasConclusion   bool `json:"has_conclusion"`
}

type Readability struct {
	Flesch          float64 `json:"flesch"`
	FernandezHuerta float64 `json:"fernandez_huerta"`
	Label           string  `json:"label"`
}

type AnalyzeResponse struct {
	WordCount          int                 `json:"word_count"`
	SentenceCount      int                 `json:"sentence_count"`
	SentenceLength     SentenceLength      `json:"sentence_length"`
	VocabularyRichness float64             `json:"vocabulary_richness"`
	Paragraphs         Paragraphs          `json:"paragraphs"`
	Readability        Readability         `json:"readability"`
	FillerWords        int                 `json:"filler_words"`
	Keywords           []string            `json:"keywords"`
	Requirements       []RequirementResult `json:"requirements"`
	Score              int                 `json:"score"`
	Feedback           []string            `json:"feedback"`
}

func getAnalyzerURL() string {
	url := os.Getenv("ANALYZER_URL")
	if url == "" {
		return "http://localhost:8001"
	}
	return url
}

func AnalyzeText(text string, requirements []string) (*AnalyzeResponse, error) {
	reqBody := AnalyzeRequest{
		Text:         text,
		Requirements: requirements,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(
		getAnalyzerURL()+"/analyze/text",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call analyzer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analyzer returned status %d", resp.StatusCode)
	}

	var result AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode analyzer response: %w", err)
	}

	return &result, nil
}
