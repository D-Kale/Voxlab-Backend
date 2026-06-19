package repositories

import (
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

type ModuleRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) *ModuleRepository {
	return &ModuleRepository{db: db}
}

func (r *ModuleRepository) FindAllByTrack(trackID int) ([]models.Module, error) {
	var modules []models.Module
	err := r.db.Where("track_id = ?", trackID).Preload("Lessons.Lesson.LessonExercises.Exercise").Order("order_index asc").Find(&modules).Error
	return modules, err
}

func (r *ModuleRepository) FindByID(id int) (*models.Module, error) {
	var module models.Module
	err := r.db.Preload("Lessons.Lesson.LessonExercises.Exercise").First(&module, id).Error
	if err != nil {
		return nil, err
	}
	return &module, nil
}

func (r *ModuleRepository) Create(module *models.Module) error {
	return r.db.Create(module).Error
}

func (r *ModuleRepository) Update(module *models.Module) error {
	return r.db.Save(module).Error
}

func (r *ModuleRepository) Delete(id int) error {
	return r.db.Delete(&models.Module{}, id).Error
}

func (r *ModuleRepository) LinkLesson(moduleID, lessonID, orderIndex int) error {
	link := models.ModuleLesson{
		ModuleID:   moduleID,
		LessonID:   lessonID,
		OrderIndex: orderIndex,
	}
	return r.db.Create(&link).Error
}
