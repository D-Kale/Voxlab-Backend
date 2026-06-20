package repositories

import (
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

type LessonExerciseRepository struct {
	db *gorm.DB
}

func NewLessonExerciseRepository(db *gorm.DB) *LessonExerciseRepository {
	return &LessonExerciseRepository{db: db}
}

func (r *LessonExerciseRepository) FindByLesson(lessonID int) ([]models.LessonExercise, error) {
	var links []models.LessonExercise
	err := r.db.Where("lesson_id = ?", lessonID).
		Preload("Exercise").
		Order("order_index asc").
		Find(&links).Error
	return links, err
}

func (r *LessonExerciseRepository) Link(lessonID int, exerciseID uuid.UUID, orderIndex int) error {
	link := models.LessonExercise{
		LessonID:   lessonID,
		ExerciseID: exerciseID,
		OrderIndex: orderIndex,
	}
	return r.db.Create(&link).Error
}

func (r *LessonExerciseRepository) Unlink(lessonID int, exerciseID uuid.UUID) error {
	return r.db.Delete(&models.LessonExercise{}, "lesson_id = ? AND exercise_id = ?", lessonID, exerciseID).Error
}

func (r *LessonExerciseRepository) UpdateOrder(lessonID int, exerciseID uuid.UUID, newOrder int) error {
	return r.db.Model(&models.LessonExercise{}).
		Where("lesson_id = ? AND exercise_id = ?", lessonID, exerciseID).
		Update("order_index", newOrder).Error
}

func (r *LessonExerciseRepository) GetNextOrderIndex(lessonID int) (int, error) {
	var max int
	err := r.db.Model(&models.LessonExercise{}).
		Select("COALESCE(MAX(order_index), 0)").
		Where("lesson_id = ?", lessonID).
		Scan(&max).Error
	return max + 1, err
}

func (r *LessonExerciseRepository) BatchReorder(lessonID int, items []models.LessonExercise) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			err := tx.Model(&models.LessonExercise{}).
				Where("lesson_id = ? AND exercise_id = ?", lessonID, item.ExerciseID).
				Update("order_index", item.OrderIndex).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *LessonExerciseRepository) FindByExercise(exerciseID uuid.UUID) ([]models.LessonExercise, error) {
	var links []models.LessonExercise
	err := r.db.Where("exercise_id = ?", exerciseID).
		Preload("Lesson").
		Order("lesson_id asc").
		Find(&links).Error
	return links, err
}
