package models

import (
	"time"

	"github.com/google/uuid"
)

type ExerciseAttempt struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key" json:"id" swaggertype:"string"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id" swaggertype:"string"`
	ExerciseID   uuid.UUID  `gorm:"type:uuid;not null" json:"exercise_id" swaggertype:"string"`
	LessonID     int        `gorm:"not null" json:"lesson_id"`
	Score        int        `gorm:"default:0" json:"score"`
	Passed       bool       `gorm:"not null" json:"passed"`
	ConsumedLife bool       `gorm:"default:false" json:"consumed_life"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
}
