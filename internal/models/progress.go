package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CompletedExercise struct {
	ExerciseID  string     `json:"exercise_id"`
	Score       int        `json:"score"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type UserProgress struct {
	UserID             uuid.UUID       `gorm:"type:uuid;primary_key" json:"user_id" swaggertype:"string"`
	LessonID           int             `gorm:"primary_key" json:"lesson_id"`
	Status             string          `gorm:"type:varchar(20);default:'in_progress'" json:"status"`
	XPEarned           int             `gorm:"default:0" json:"xp_earned"`
	CompletedExercises json.RawMessage `gorm:"type:jsonb;default:'[]'::jsonb" json:"completed_exercises"`
	CompletedAt        *time.Time      `json:"completed_at,omitempty"`
	CreatedAt          time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserProgress) TableName() string { return "user_progress" }
