package services

import (
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type ExerciseService struct {
	repo *repositories.ExerciseRepository
}

func NewExerciseService(repo *repositories.ExerciseRepository) *ExerciseService {
	return &ExerciseService{repo: repo}
}

func (s *ExerciseService) GetAllByLesson(lessonID int) ([]models.Exercise, error) {
	return s.repo.FindAllByLesson(lessonID)
}

func (s *ExerciseService) GetByID(id uuid.UUID) (*models.Exercise, error) {
	return s.repo.FindByID(id)
}

func (s *ExerciseService) Create(exercise *models.Exercise) error {
	return s.repo.Create(exercise)
}

func (s *ExerciseService) Update(exercise *models.Exercise) error {
	return s.repo.Update(exercise)
}

func (s *ExerciseService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}
