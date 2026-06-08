package services

import (
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type EducationalService struct {
	repo *repositories.EducationalRepository
}

func NewEducationalService(repo *repositories.EducationalRepository) *EducationalService {
	return &EducationalService{repo: repo}
}

func (s *EducationalService) GetTracks() ([]models.Track, error) {
	return s.repo.FindAllTracks()
}
