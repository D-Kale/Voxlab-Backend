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

func (r *ModuleRepository) UnlinkLesson(moduleID, lessonID int) error {
	return r.db.Delete(&models.ModuleLesson{}, "module_id = ? AND lesson_id = ?", moduleID, lessonID).Error
}

type ModuleOrderItem struct {
	ID         int `json:"id"`
	OrderIndex int `json:"order_index"`
}

func (r *ModuleRepository) BatchUpdateOrder(items []ModuleOrderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Model(&models.Module{}).Where("id = ?", item.ID).Update("order_index", item.OrderIndex).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ModuleRepository) ReorderLessons(moduleID int, items []models.ModuleLesson) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			err := tx.Model(&models.ModuleLesson{}).
				Where("module_id = ? AND lesson_id = ?", moduleID, item.LessonID).
				Update("order_index", item.OrderIndex).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}
