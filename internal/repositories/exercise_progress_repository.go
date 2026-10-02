package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExerciseProgressRepository struct {
	db *gorm.DB
}

func NewExerciseProgressRepository(db *gorm.DB) *ExerciseProgressRepository {
	return &ExerciseProgressRepository{db: db}
}

func (r *ExerciseProgressRepository) UpsertExerciseProgress(progress *models.UserExerciseProgress) error {
	now := time.Now()
	progress.UpdatedAt = now

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "exercise_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "completed_at", "updated_at", "lesson_id"}),
	}).Create(progress).Error
}

func (r *ExerciseProgressRepository) FindByUserAndExercise(userID uuid.UUID, exerciseID string) (*models.UserExerciseProgress, error) {
	var progress models.UserExerciseProgress
	err := r.db.Where("user_id = ? AND exercise_id = ?", userID, exerciseID).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *ExerciseProgressRepository) FindAllByUser(userID uuid.UUID) ([]models.UserExerciseProgress, error) {
	var progress []models.UserExerciseProgress
	err := r.db.Where("user_id = ?", userID).Order("completed_at asc").Find(&progress).Error
	return progress, err
}

func (r *ExerciseProgressRepository) FindCompletedExercises(userID uuid.UUID) ([]string, error) {
	var exerciseIDs []string
	err := r.db.Model(&models.UserExerciseProgress{}).
		Where("user_id = ?", userID).
		Pluck("exercise_id", &exerciseIDs).Error
	return exerciseIDs, err
}

type WeeklyChartPoint struct {
	Date string
	XP   int
}

func (r *ExerciseProgressRepository) GetWeeklyChart(userID uuid.UUID, days int) ([]WeeklyChartPoint, error) {
	var results []WeeklyChartPoint
	cutoff := time.Now().AddDate(0, 0, -days)

	err := r.db.Raw(`
		SELECT 
			to_char(completed_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') as date,
			SUM(score) as xp
		FROM user_exercise_progress
		WHERE user_id = ? AND completed_at >= ?
		GROUP BY to_char(completed_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')
		ORDER BY date ASC
	`, userID, cutoff).Scan(&results).Error

	return results, err
}