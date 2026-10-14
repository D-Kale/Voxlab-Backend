package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// fakeAnalyzer stands in for the Python service and records what the API
// forwarded to it.
func fakeAnalyzer(t *testing.T, status int, body string) (*httptest.Server, *http.Request, *[]byte) {
	t.Helper()
	var got http.Request
	var audio []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/analyze/speech" {
			t.Errorf("unexpected analyzer path %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("analyzer got invalid multipart: %v", err)
		}
		got = *r
		if f, _, err := r.FormFile("audio"); err == nil {
			audio, _ = io.ReadAll(f)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("ANALYZER_URL", srv.URL)
	return srv, &got, &audio
}

func speechRequest(t *testing.T, audio []byte, fields map[string]string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if audio != nil {
		part, _ := w.CreateFormFile("audio", "take.m4a")
		_, _ = part.Write(audio)
	}
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPost, "/analyze-speech", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func serve(req *http.Request) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/analyze-speech", (&ExerciseController{}).AnalyzeSpeech)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestAnalyzeSpeechForwardsAudioAndTargets(t *testing.T) {
	_, got, audio := fakeAnalyzer(t, http.StatusOK, `{"transcript":"hola a todos","words_per_minute":130,"filler_words":2,"score":82,"score_breakdown":{"ritmo":{"score":100,"weight":0.25,"feedbacks":[]}},"feedback":["Buen ritmo"]}`)

	rec := serve(speechRequest(t, []byte("fake-aac"), map[string]string{
		"min_seconds":    "20",
		"target_seconds": "60",
	}))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if string(*audio) != "fake-aac" {
		t.Errorf("forwarded audio = %q", *audio)
	}
	if got.FormValue("min_seconds") != "20" || got.FormValue("target_seconds") != "60" {
		t.Errorf("forwarded targets = %q / %q", got.FormValue("min_seconds"), got.FormValue("target_seconds"))
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Transcript  string `json:"transcript"`
			Score       int    `json:"score"`
			FillerWords int    `json:"filler_words"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || body.Data.Score != 82 || body.Data.FillerWords != 2 || body.Data.Transcript != "hola a todos" {
		t.Errorf("unexpected envelope: %+v", body)
	}
}

func TestAnalyzeSpeechRequiresAudio(t *testing.T) {
	rec := serve(speechRequest(t, nil, nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAnalyzeSpeechRejectsInvalidTargets(t *testing.T) {
	rec := serve(speechRequest(t, []byte("x"), map[string]string{"min_seconds": "abc"}))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAnalyzeSpeechMapsUnreadableAudioTo422(t *testing.T) {
	fakeAnalyzer(t, http.StatusUnprocessableEntity, `{"detail":"No se pudo procesar el audio"}`)
	rec := serve(speechRequest(t, []byte("not audio"), nil))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body = %s", rec.Code, rec.Body.String())
	}
}

func TestAnalyzeSpeechForwardsTopicAndAiFeedback(t *testing.T) {
	_, got, _ := fakeAnalyzer(t, http.StatusOK, `{"transcript":"hola","score":70,"score_breakdown":{},"feedback":[],"ai_feedback":"Cerrá con una idea fuerte.","ai_model":"google/gemma-4-31b-it:free"}`)

	rec := serve(speechRequest(t, []byte("fake-aac"), map[string]string{"topic": "  Presentación breve  "}))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got.FormValue("topic") != "Presentación breve" {
		t.Errorf("forwarded topic = %q", got.FormValue("topic"))
	}
	var body struct {
		Data struct {
			AiFeedback *string `json:"ai_feedback"`
			AiModel    *string `json:"ai_model"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.AiFeedback == nil || *body.Data.AiFeedback != "Cerrá con una idea fuerte." {
		t.Errorf("ai_feedback = %v", body.Data.AiFeedback)
	}
}

func TestAnalyzeSpeechMapsBusyAnalyzerTo503(t *testing.T) {
	fakeAnalyzer(t, http.StatusServiceUnavailable, `{"detail":"ocupado"}`)
	rec := serve(speechRequest(t, []byte("audio"), nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body = %s", rec.Code, rec.Body.String())
	}
}
