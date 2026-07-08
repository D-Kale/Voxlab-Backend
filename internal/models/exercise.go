package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ExerciseType string

const (
	ExerciseTypeReading        ExerciseType = "reading"
	ExerciseTypeQuiz           ExerciseType = "quiz"
	ExerciseTypeAudio          ExerciseType = "audio"
	ExerciseTypeOratoryMinigame ExerciseType = "oratory_minigame"
	ExerciseTypeVideo          ExerciseType = "video"
	ExerciseTypeWriting        ExerciseType = "writing"
)

// Exercise represents a polymorphic learning exercise that can be reused
// across multiple lessons via the LessonExercise pivot table.
// The Content field is JSONB and its structure depends on the Type field.
// Supported types: quiz, reading, oratory_minigame, audio, video, writing.
type Exercise struct {
	ID            uuid.UUID       `gorm:"type:uuid;primary_key" json:"id" swaggertype:"string"`
	Name          string          `gorm:"type:varchar(255);default:''" json:"name" example:"Quiz de liderazgo"`
	Type          ExerciseType    `gorm:"type:varchar(50);not null" json:"type"`
	Content       json.RawMessage `gorm:"type:jsonb;not null" json:"content" swaggertype:"object"`
	PassingScore  int             `gorm:"type:int;default:60" json:"passing_score" example:"60"`
	CreatedAt     time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// QuizQuestion represents a single multiple-choice question.
type QuizQuestion struct {
	Question     string   `json:"question" example:"What is public speaking?"`
	Options      []string `json:"options" example:"Option A,Option B,Option C,Option D"`
	CorrectIndex int      `json:"correct_index" example:"0"`
	Explanation  string   `json:"explanation,omitempty" example:"Option A is correct because..."`
}

// ExerciseContentQuiz represents the JSONB content for a quiz exercise.
type ExerciseContentQuiz struct {
	Questions         []QuizQuestion `json:"questions"`
	PointsPerQuestion int            `json:"points_per_question" example:"10"`
}

// ExerciseContentOratoryMinigame represents the JSONB content for an oratory minigame.
type ExerciseContentOratoryMinigame struct {
	Prompt             string   `json:"prompt" example:"Record a 30-second speech about leadership"`
	Topic              string   `json:"topic" example:"Leadership"`
	DurationSeconds    int      `json:"duration_seconds" example:"30"`
	MinDurationSeconds int      `json:"min_duration_seconds" example:"15"`
	Requirements       []string `json:"requirements" example:"Clear introduction,Use at least 3 key points,Strong conclusion"`
	Points             int      `json:"points" example:"20"`
}

// ExerciseContentReading represents the JSONB content for a reading exercise.
type ExerciseContentReading struct {
	Title       string `json:"title" example:"The Art of Speech"`
	Content     string `json:"content" example:"Full reading text here..."`
	ReadingTime int    `json:"reading_time_seconds" example:"120"`
	Points      int    `json:"points" example:"5"`
}

// ExerciseContentWriting represents the JSONB content for a writing exercise.
type ExerciseContentWriting struct {
	Prompt       string   `json:"prompt" example:"Write a 200-word essay about leadership"`
	MinWords     int      `json:"min_words" example:"100"`
	MaxWords     int      `json:"max_words" example:"500"`
	Requirements []string `json:"requirements" example:"Include a thesis,Support with examples"`
	Points       int      `json:"points" example:"20"`
}

// ExerciseContentAudio represents the JSONB content for an audio exercise.
type ExerciseContentAudio struct {
	Prompt          string `json:"prompt" example:"Read this paragraph aloud..."`
	DurationSeconds int    `json:"duration_seconds" example:"60"`
	Points          int    `json:"points" example:"15"`
}

// ExerciseContentVideo represents the JSONB content for a video exercise.
type ExerciseContentVideo struct {
	Prompt          string `json:"prompt" example:"Record a video introducing yourself..."`
	DurationSeconds int    `json:"duration_seconds" example:"120"`
	Points          int    `json:"points" example:"25"`
}
