package services

import (
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type CommunityService struct {
	repo *repositories.CommunityRepository
}

func NewCommunityService(repo *repositories.CommunityRepository) *CommunityService {
	return &CommunityService{repo: repo}
}

func (s *CommunityService) CreateReaction(reaction *models.UserReaction) error {
	return s.repo.Create(reaction)
}
