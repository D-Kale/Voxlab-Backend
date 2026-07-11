package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

type ExerciseAttemptRepository struct {
	db *gorm.DB
}

func NewExerciseAttemptRepository(db *gorm.DB) *ExerciseAttemptRepository {
	return &ExerciseAttemptRepository{db: db}
}

func (r *ExerciseAttemptRepository) Create(attempt *models.ExerciseAttempt) error {
	return r.db.Create(attempt).Error
}

func (r *ExerciseAttemptRepository) FindByUserAndExercise(userID, exerciseID uuid.UUID) ([]models.ExerciseAttempt, error) {
	var attempts []models.ExerciseAttempt
	err := r.db.Where("user_id = ? AND exercise_id = ?", userID, exerciseID).Order("created_at desc").Find(&attempts).Error
	return attempts, err
}

func (r *ExerciseAttemptRepository) FindByUserAndLesson(userID uuid.UUID, lessonID int) ([]models.ExerciseAttempt, error) {
	var attempts []models.ExerciseAttempt
	err := r.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).Order("created_at desc").Find(&attempts).Error
	return attempts, err
}

type DailyAttemptCount struct {
	Date  time.Time
	Count int
}

func (r *ExerciseAttemptRepository) CountPassedAttemptsThisWeek(userID uuid.UUID) ([]DailyAttemptCount, error) {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	weekStart := now.AddDate(0, 0, -int(weekday-time.Monday))

	var results []struct {
		Date  time.Time `gorm:"column:date"`
		Count int       `gorm:"column:count"`
	}

	err := r.db.Raw(`
		SELECT
			DATE(created_at) as date,
			COUNT(*) as count
		FROM exercise_attempts
		WHERE user_id = ?
			AND passed = true
			AND created_at >= ?
			AND created_at < ?
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, userID, weekStart, weekStart.AddDate(0, 0, 7)).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make([]DailyAttemptCount, len(results))
	for i, r := range results {
		counts[i] = DailyAttemptCount{Date: r.Date, Count: r.Count}
	}
	return counts, nil
}
