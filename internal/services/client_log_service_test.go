package services

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/voxlab/voxlab-backend/internal/models"
)

type fakeRepo struct {
	created []*models.ClientLog
	err     error
}

func (f *fakeRepo) Create(l *models.ClientLog) error {
	if f.err != nil {
		return f.err
	}
	f.created = append(f.created, l)
	return nil
}

func TestRecordPersistsAndMapsUserID(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewClientLogService(repo)

	svc.Record(ClientLogInput{
		Level:   "error",
		Message: "boom",
		Stack:   "Error: boom\n  at foo",
	}, "user-123")

	if len(repo.created) != 1 {
		t.Fatalf("expected 1 log persisted, got %d", len(repo.created))
	}

	got := repo.created[0]
	if got.UserID != "user-123" {
		t.Errorf("UserID = %q, want user-123", got.UserID)
	}
	if got.Level != "error" {
		t.Errorf("Level = %q, want error", got.Level)
	}
	if got.Stack == "" {
		t.Error("Stack should be preserved")
	}
}

func TestRecordDoesNotFailWhenRepoFails(t *testing.T) {
	// La telemetría no debe romper la request: si el insert falla, el error se
	// loguea pero Record no lo retorna.
	repo := &fakeRepo{err: errors.New("db down")}
	svc := NewClientLogService(repo)

	svc.Record(ClientLogInput{Level: "error", Message: "boom"}, "u1")
}

func TestRecordNormalizesLevels(t *testing.T) {
	cases := map[string]string{
		"WARNING": "warn",
		"Error":   "error",
		"":        "error",
		"bogus":   "info",
		"warn":    "warn",
	}

	for input, want := range cases {
		repo := &fakeRepo{}
		svc := NewClientLogService(repo)
		svc.Record(ClientLogInput{Level: input, Message: "m"}, "u")

		if got := repo.created[0].Level; got != want {
			t.Errorf("level %q -> %q, want %q", input, got, want)
		}
	}
}

func TestRecordDropsInvalidAndOversizedContext(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewClientLogService(repo)

	svc.Record(ClientLogInput{
		Level:   "error",
		Message: "m",
		Context: json.RawMessage(`{not json`),
	}, "u")
	if got := string(repo.created[0].Context); got != `{"invalid":true}` {
		t.Errorf("invalid context = %q, want {\"invalid\":true}", got)
	}

	repo2 := &fakeRepo{}
	svc2 := NewClientLogService(repo2)
	huge := append(json.RawMessage(`{"a":"`), []byte(strings.Repeat("x", maxContextBytes))...)
	huge = append(huge, []byte(`"}`)...)
	svc2.Record(ClientLogInput{Level: "error", Message: "m", Context: huge}, "u")

	if got := string(repo2.created[0].Context); got != `{"truncated":true}` {
		t.Errorf("oversized context = %q, want truncated marker", got)
	}
}

func TestRecordTruncatesLongFields(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewClientLogService(repo)

	svc.Record(ClientLogInput{
		Level:   "error",
		Message: strings.Repeat("m", 9000),
	}, "u")

	if got := len(repo.created[0].Message); got != 4000 {
		t.Errorf("message length = %d, want 4000", got)
	}
}
