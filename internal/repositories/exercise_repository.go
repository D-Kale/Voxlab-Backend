package repositories

import (
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

type ExerciseRepository struct {
	db *gorm.DB
}

func NewExerciseRepository(db *gorm.DB) *ExerciseRepository {
	return &ExerciseRepository{db: db}
}

func (r *ExerciseRepository) FindAllByLesson(lessonID int) ([]models.Exercise, error) {
	var exercises []models.Exercise
	err := r.db.Where("lesson_id = ?", lessonID).Order("order_index asc").Find(&exercises).Error
	return exercises, err
}

func (r *ExerciseRepository) FindByID(id uuid.UUID) (*models.Exercise, error) {
	var exercise models.Exercise
	err := r.db.First(&exercise, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (r *ExerciseRepository) Create(exercise *models.Exercise) error {
	return r.db.Create(exercise).Error
}

func (r *ExerciseRepository) Update(exercise *models.Exercise) error {
	return r.db.Save(exercise).Error
}

func (r *ExerciseRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Exercise{}, "id = ?", id).Error
}
