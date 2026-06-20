package repositories

import (
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

type LessonRepository struct {
	db *gorm.DB
}

func NewLessonRepository(db *gorm.DB) *LessonRepository {
	return &LessonRepository{db: db}
}

func (r *LessonRepository) FindAllByModule(moduleID int) ([]models.ModuleLesson, error) {
	var links []models.ModuleLesson
	err := r.db.Where("module_id = ?", moduleID).
		Preload("Lesson.LessonExercises.Exercise").
		Order("order_index asc").
		Find(&links).Error
	return links, err
}

func (r *LessonRepository) FindByID(id int) (*models.Lesson, error) {
	var lesson models.Lesson
	err := r.db.Preload("LessonExercises.Exercise").First(&lesson, id).Error
	if err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (r *LessonRepository) Create(lesson *models.Lesson) error {
	return r.db.Create(lesson).Error
}

func (r *LessonRepository) Update(lesson *models.Lesson) error {
	return r.db.Save(lesson).Error
}

func (r *LessonRepository) Delete(id int) error {
	return r.db.Delete(&models.Lesson{}, id).Error
}

func (r *LessonRepository) FindAll() ([]models.Lesson, error) {
	var lessons []models.Lesson
	err := r.db.Order("title asc").Find(&lessons).Error
	return lessons, err
}

func (r *LessonRepository) FindModulesByLesson(lessonID int) ([]models.ModuleLesson, error) {
	var links []models.ModuleLesson
	err := r.db.Where("lesson_id = ?", lessonID).
		Preload("Module").
		Order("order_index asc").
		Find(&links).Error
	return links, err
}

func (r *LessonRepository) FindExercisesByLesson(lessonID int) ([]models.LessonExercise, error) {
	var links []models.LessonExercise
	err := r.db.Where("lesson_id = ?", lessonID).
		Preload("Exercise").
		Order("order_index asc").
		Find(&links).Error
	return links, err
}
