package services

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/logger"
	"github.com/voxlab/voxlab-backend/internal/models"
)

// maxContextBytes evita que un contexto enorme (por ejemplo un dump completo
// de un response) infle la base. El frontend trunca antes, esto es el backstop.
const maxContextBytes = 8 * 1024

// ClientLogInput es la entrada aceptada por POST /api/v1/logs.
type ClientLogInput struct {
	Level     string          `json:"level"`
	Message   string          `json:"message"`
	Source    string          `json:"source"`
	URL       string          `json:"url"`
	Stack     string          `json:"stack"`
	RequestID string          `json:"request_id"`
	UserAgent string          `json:"user_agent"`
	Context   json.RawMessage `json:"context"`
}

type ClientLogService struct {
	repo ClientLogWriter
	now  func() time.Time
}

// ClientLogWriter abstrae la persistencia para poder testear sin DB.
type ClientLogWriter interface {
	Create(log *models.ClientLog) error
}

func NewClientLogService(repo ClientLogWriter) *ClientLogService {
	return &ClientLogService{repo: repo, now: time.Now}
}

var validLevels = map[string]bool{
	"error":   true,
	"warn":    true,
	"warning": true,
	"info":    true,
	"debug":   true,
}

// Record persiste un log del cliente y lo escribe en el log del servidor.
//
// Devolver un error de persistencia NO se propaga al endpoint: si el insert
// falla igual se puede haber perdido la fila, pero el usuario ya tiene su
// toast y no debería ver un 500 por un error de telemetría.
func (s *ClientLogService) Record(input ClientLogInput, userID string) {
	level := normalizeLevel(input.Level)

	entry := &models.ClientLog{
		// GORM no genera los UUID de este proyecto: se asignan a mano en
		// cada service (ver auth_service y attempt_service).
		ID:        uuid.New(),
		Level:     level,
		Message:   truncate(input.Message, 4000),
		Source:    truncate(input.Source, 255),
		URL:       truncate(input.URL, 512),
		Stack:     truncate(input.Stack, 8000),
		RequestID: truncate(input.RequestID, 64),
		UserAgent: truncate(input.UserAgent, 512),
		UserID:    truncate(userID, 64),
		Context:   truncateContext(input.Context),
	}

	// Escribir en el log del servidor es lo que permite ver la falla en
	// `docker compose logs backend` sin tocar la base.
	logArgs := []any{
		"origin", "client",
		"request_id", entry.RequestID,
		"user_id", entry.UserID,
		"url", entry.URL,
		"source", entry.Source,
	}
	if entry.Stack != "" {
		logArgs = append(logArgs, "stack", entry.Stack)
	}
	if len(entry.Context) > 0 {
		logArgs = append(logArgs, "context", string(entry.Context))
	}

	if level == "error" {
		logger.Error(entry.Message, logArgs...)
	} else {
		logger.Warn(entry.Message, logArgs...)
	}

	if err := s.repo.Create(entry); err != nil {
		logger.Error("failed to persist client log",
			"request_id", entry.RequestID,
			"error", err.Error(),
			"log_message", entry.Message,
		)
	}
}

func normalizeLevel(level string) string {
	l := strings.ToLower(strings.TrimSpace(level))
	if l == "" {
		return "error"
	}
	if l == "warning" {
		return "warn"
	}
	if !validLevels[l] {
		return "info"
	}
	return l
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func truncateContext(ctx json.RawMessage) json.RawMessage {
	if len(ctx) == 0 {
		return nil
	}
	if !json.Valid(ctx) {
		// No dejar que un payload malformado rompa el insert.
		return json.RawMessage(`{"invalid":true}`)
	}
	if len(ctx) > maxContextBytes {
		return json.RawMessage(`{"truncated":true}`)
	}
	return ctx
}
