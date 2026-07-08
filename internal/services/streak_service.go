package services

import (
	"errors"
	"time"

	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

var errNoLivesToRecover = errors.New("no lives available to recover streak")

type StreakService struct {
	userRepo *repositories.UserRepository
}

func NewStreakService(userRepo *repositories.UserRepository) *StreakService {
	return &StreakService{userRepo: userRepo}
}

func (s *StreakService) IsStreakAtRisk(user *models.User) bool {
	if user.LastActivityAt == nil {
		return false
	}
	elapsed := time.Since(*user.LastActivityAt)
	return elapsed >= time.Duration(StreakGraceHours)*time.Hour
}

func (s *StreakService) IsStreakLost(user *models.User) bool {
	if user.LastActivityAt == nil {
		return false
	}
	elapsed := time.Since(*user.LastActivityAt)
	return elapsed >= time.Duration(StreakGraceHours+StreakRecoveryHours)*time.Hour
}

func (s *StreakService) RecoverStreak(user *models.User) error {
	if user.Lives <= 0 {
		return errNoLivesToRecover
	}

	user.Lives--
	now := time.Now()
	user.LastActivityAt = &now
	return s.userRepo.Update(user)
}

func (s *StreakService) TouchActivity(user *models.User) {
	now := time.Now()
	user.LastActivityAt = &now
}
