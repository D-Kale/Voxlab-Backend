package models

import (
	"time"

	"github.com/google/uuid"
)

type UserExerciseProgress struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key" json:"id" swaggertype:"string"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id" swaggertype:"string"`
	ExerciseID   string     `gorm:"type:varchar(64);not null" json:"exercise_id"`
	LessonID     *int       `json:"lesson_id,omitempty"`
	Score        int        `gorm:"not null" json:"score"`
	CompletedAt  time.Time  `gorm:"type:timestamptz;not null" json:"completed_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

type WeeklyChartPoint struct {
	Date string `json:"date"`
	XP   int    `json:"xp"`
}

type UserProgressResponse struct {
	XP                   int                 `json:"xp"`
	StreakDays           int                 `json:"streak_days"`
	CompletedExerciseIds []string            `json:"completed_exercise_ids"`
	WeeklyChart          []WeeklyChartPoint  `json:"weekly_chart"`
	LastSyncedAt         *time.Time          `json:"last_synced_at,omitempty"`
}