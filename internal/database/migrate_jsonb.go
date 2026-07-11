package database

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type completedExerciseV1 struct {
	ExerciseID string `json:"exercise_id"`
	Score      int    `json:"score"`
}

type completedExerciseV2 struct {
	ExerciseID  string     `json:"exercise_id"`
	Score       int        `json:"score"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type userProgressRow struct {
	UserID             uuid.UUID
	LessonID           int
	CompletedExercises []byte `gorm:"column:completed_exercises"`
	UpdatedAt          time.Time
}

// MigrateJSONBCompletedAt adds completed_at to existing completed_exercises JSONB entries
// that were created before the field existed.
func MigrateJSONBCompletedAt() error {
	log.Println("Running JSONB completed_at migration...")

	var rows []userProgressRow
	if err := DB.Table("user_progress").
		Select("user_id, lesson_id, completed_exercises, updated_at").
		Where("completed_exercises IS NOT NULL").
		Where("completed_exercises != '[]'::jsonb").
		Find(&rows).Error; err != nil {
		return fmt.Errorf("reading progress rows: %w", err)
	}

	migrated := 0
	for _, row := range rows {
		// Try parsing as V1 first
		var v1 []completedExerciseV1
		if err := json.Unmarshal(row.CompletedExercises, &v1); err != nil {
			continue
		}
		if len(v1) == 0 {
			continue
		}

		// Check if any entry lacks completed_at by trying V2 parse
		var v2 []completedExerciseV2
		if err := json.Unmarshal(row.CompletedExercises, &v2); err != nil {
			continue
		}

		needsMigration := false
		for _, ex := range v2 {
			if ex.CompletedAt == nil {
				needsMigration = true
				break
			}
		}
		if !needsMigration {
			continue
		}

		// Build V2 with completed_at from updated_at
		migratedV2 := make([]completedExerciseV2, len(v1))
		for i, ex := range v1 {
			migratedV2[i] = completedExerciseV2{
				ExerciseID:  ex.ExerciseID,
				Score:       ex.Score,
				CompletedAt: &row.UpdatedAt,
			}
		}

		data, err := json.Marshal(migratedV2)
		if err != nil {
			log.Printf("Error marshaling row %s/%d: %v", row.UserID, row.LessonID, err)
			continue
		}

		if err := DB.Table("user_progress").
			Where("user_id = ? AND lesson_id = ?", row.UserID, row.LessonID).
			UpdateColumn("completed_exercises", data).Error; err != nil {
			log.Printf("Error updating row %s/%d: %v", row.UserID, row.LessonID, err)
			continue
		}
		migrated++
	}

	log.Printf("JSONB completed_at migration: %d rows updated", migrated)
	return nil
}
