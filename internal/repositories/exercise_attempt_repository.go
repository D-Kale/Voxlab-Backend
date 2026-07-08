package repositories

import (
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
