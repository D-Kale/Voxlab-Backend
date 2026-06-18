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
	MinWords     *int     `json:"min_words,omitempty"`
	MaxWords     *int     `json:"max_words,omitempty"`
}

type ScoreBreakdownItem struct {
	Score     int      `json:"score"`
	Weight    float64  `json:"weight"`
	Feedbacks []string `json:"feedbacks"`
}

type SentenceAnalysis struct {
	SentenceCount  int     `json:"sentence_count"`
	AvgLength      float64 `json:"avg_length"`
	StdLength      float64 `json:"std_length"`
	ConnectorRatio float64 `json:"connector_ratio"`
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
	WordCount           int                            `json:"word_count"`
	SentenceCount       int                            `json:"sentence_count"`
	SentenceLength      SentenceLength                 `json:"sentence_length"`
	SentenceAnalysis    SentenceAnalysis               `json:"sentence_analysis"`
	Paragraphs          Paragraphs                     `json:"paragraphs"`
	VocabularyRichness  float64                        `json:"vocabulary_richness"`
	OovRatio            float64                        `json:"oov_ratio"`
	Readability         Readability                    `json:"readability"`
	FillerWords         int                            `json:"filler_words"`
	Keywords            []string                       `json:"keywords"`
	Requirements        []RequirementResult            `json:"requirements"`
	GibberishDetected   bool                           `json:"gibberish_detected"`
	Score               int                            `json:"score"`
	ScoreBreakdown      map[string]ScoreBreakdownItem  `json:"score_breakdown"`
	Feedback            []string                       `json:"feedback"`
}

func analyzerURLs() []string {
	primary := os.Getenv("ANALYZER_URL")
	if primary != "" {
		return []string{primary, "http://localhost:8001"}
	}
	return []string{"http://localhost:8001"}
}

func AnalyzeText(text string, requirements []string, minWords, maxWords *int) (*AnalyzeResponse, error) {
	reqBody := AnalyzeRequest{
		Text:         text,
		Requirements: requirements,
		MinWords:     minWords,
		MaxWords:     maxWords,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 180 * time.Second}
	urls := analyzerURLs()
	var lastErr error

	for _, baseURL := range urls {
		resp, err := client.Post(
			baseURL+"/analyze/text",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("analyzer returned status %d", resp.StatusCode)
			continue
		}

		var result AnalyzeResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("failed to decode analyzer response: %w", err)
			continue
		}

		resp.Body.Close()
		return &result, nil
	}

	return nil, fmt.Errorf("analyzer unreachable: %w", lastErr)
}
