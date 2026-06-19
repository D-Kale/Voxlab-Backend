package models

import (
	"github.com/google/uuid"
)

// LessonExercise is the pivot table linking exercises to lessons (many-to-many).
// Each link has its own order_index so the same exercise can appear at different
// positions in different lessons. This allows exercise reuse across the curriculum
// while maintaining per-lesson ordering.
type LessonExercise struct {
	LessonID   int       `gorm:"primaryKey" json:"lesson_id"`
	ExerciseID uuid.UUID `gorm:"primaryKey;type:uuid" json:"exercise_id" swaggertype:"string"`
	OrderIndex int       `gorm:"not null;default:1" json:"order_index"`
	Lesson     Lesson    `gorm:"foreignKey:LessonID" json:"-"`
	Exercise   Exercise  `gorm:"foreignKey:ExerciseID" json:"exercise"`
}
