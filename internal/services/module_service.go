package services

import (
	"fmt"

	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type ModuleService struct {
	repo *repositories.ModuleRepository
}

func NewModuleService(repo *repositories.ModuleRepository) *ModuleService {
	return &ModuleService{repo: repo}
}

func (s *ModuleService) GetAllByTrack(trackID int) ([]models.Module, error) {
	return s.repo.FindAllByTrack(trackID)
}

func (s *ModuleService) GetByID(id int) (*models.Module, error) {
	return s.repo.FindByID(id)
}

func (s *ModuleService) Create(module *models.Module) error {
	return s.repo.Create(module)
}

func (s *ModuleService) Update(module *models.Module) error {
	return s.repo.Update(module)
}

func (s *ModuleService) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *ModuleService) LinkLesson(moduleID, lessonID int) error {
	if moduleID <= 0 || lessonID <= 0 {
		return fmt.Errorf("module_id and lesson_id must be positive integers")
	}
	return s.repo.LinkLesson(moduleID, lessonID, 0)
}
