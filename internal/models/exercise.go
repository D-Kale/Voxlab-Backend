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

type Exercise struct {
	ID         uuid.UUID       `gorm:"type:uuid;primary_key" json:"id" swaggertype:"string"`
	LessonID   int             `gorm:"not null" json:"lesson_id"`
	Type       ExerciseType    `gorm:"type:varchar(50);not null" json:"type"`
	OrderIndex int             `gorm:"not null" json:"order_index"`
	Content    json.RawMessage `gorm:"type:jsonb;not null" json:"content"`
	Lesson     Lesson          `gorm:"foreignKey:LessonID" json:"-"`
	CreatedAt  time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

type QuizQuestion struct {
	Question     string   `json:"question"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
	Explanation  string   `json:"explanation,omitempty"`
}

type ExerciseContentQuiz struct {
	Questions          []QuizQuestion `json:"questions"`
	PointsPerQuestion  int            `json:"points_per_question"`
}

type ExerciseContentOratoryMinigame struct {
	Prompt             string   `json:"prompt"`
	Topic              string   `json:"topic"`
	DurationSeconds    int      `json:"duration_seconds"`
	MinDurationSeconds int      `json:"min_duration_seconds"`
	Requirements       []string `json:"requirements"`
	Points             int      `json:"points"`
}

type ExerciseContentReading struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	ReadingTime int    `json:"reading_time_seconds"`
	Points      int    `json:"points"`
}

type ExerciseContentWriting struct {
	Prompt       string   `json:"prompt"`
	MinWords     int      `json:"min_words"`
	MaxWords     int      `json:"max_words"`
	Requirements []string `json:"requirements"`
	Points       int      `json:"points"`
}
