package services

import (
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type TrackService struct {
	repo *repositories.TrackRepository
}

func NewTrackService(repo *repositories.TrackRepository) *TrackService {
	return &TrackService{repo: repo}
}

func (s *TrackService) GetAll() ([]models.Track, error) {
	return s.repo.FindAll()
}

func (s *TrackService) GetByID(id int) (*models.Track, error) {
	return s.repo.FindByID(id)
}

func (s *TrackService) Create(track *models.Track) error {
	return s.repo.Create(track)
}

func (s *TrackService) Update(track *models.Track) error {
	return s.repo.Update(track)
}

func (s *TrackService) Delete(id int) error {
	return s.repo.Delete(id)
}
