package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

// SpeechAnalyzeResponse is the analyzer's /analyze/speech payload: a
// transcript plus delivery metrics (pace, pauses, filler words).
type SpeechAnalyzeResponse struct {
	Transcript          string                        `json:"transcript"`
	DurationSeconds     float64                       `json:"duration_seconds"`
	SpeakingSeconds     float64                       `json:"speaking_seconds"`
	WordCount           int                           `json:"word_count"`
	WordsPerMinute      int                           `json:"words_per_minute"`
	PauseCount          int                           `json:"pause_count"`
	LongPauseCount      int                           `json:"long_pause_count"`
	LongestPauseSeconds float64                       `json:"longest_pause_seconds"`
	FillerWords         int                           `json:"filler_words"`
	FillerBreakdown     map[string]int                `json:"filler_breakdown"`
	Score               int                           `json:"score"`
	ScoreBreakdown      map[string]ScoreBreakdownItem `json:"score_breakdown"`
	Feedback            []string                      `json:"feedback"`
}

// SpeechAnalyzeOptions are the optional targets of the oratory exercise.
type SpeechAnalyzeOptions struct {
	MinSeconds    *float64
	TargetSeconds *float64
}

// ErrUnprocessableAudio means the analyzer could read the request but not
// the audio itself (corrupt file, unsupported codec).
type ErrUnprocessableAudio struct{ Detail string }

func (e ErrUnprocessableAudio) Error() string { return e.Detail }

// AnalyzeSpeech uploads an audio take to the analyzer for transcription
// and delivery analysis. Transcription runs on CPU, so the timeout is long.
func AnalyzeSpeech(
	audio io.Reader,
	filename string,
	opts SpeechAnalyzeOptions,
) (*SpeechAnalyzeResponse, error) {
	// The body is buffered once so it can be replayed against the fallback
	// analyzer URL.
	payload, err := io.ReadAll(audio)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio: %w", err)
	}

	client := &http.Client{Timeout: 150 * time.Second}
	var lastErr error
	for _, baseURL := range analyzerURLs() {
		body, contentType, err := speechForm(payload, filename, opts)
		if err != nil {
			return nil, err
		}
		resp, err := client.Post(baseURL+"/analyze/speech", contentType, body)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusUnprocessableEntity {
			var detail struct {
				Detail string `json:"detail"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&detail)
			resp.Body.Close()
			return nil, ErrUnprocessableAudio{Detail: detail.Detail}
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("analyzer returned status %d", resp.StatusCode)
			continue
		}

		var result SpeechAnalyzeResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to decode analyzer response: %w", err)
			continue
		}
		return &result, nil
	}

	return nil, fmt.Errorf("analyzer unreachable: %w", lastErr)
}

func speechForm(
	audio []byte,
	filename string,
	opts SpeechAnalyzeOptions,
) (io.Reader, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("audio", filename)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(audio); err != nil {
		return nil, "", err
	}
	if opts.MinSeconds != nil {
		_ = w.WriteField("min_seconds", strconv.FormatFloat(*opts.MinSeconds, 'f', -1, 64))
	}
	if opts.TargetSeconds != nil {
		_ = w.WriteField("target_seconds", strconv.FormatFloat(*opts.TargetSeconds, 'f', -1, 64))
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return &buf, w.FormDataContentType(), nil
}
