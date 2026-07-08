package services

import (
	"time"

	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

const (
	MaxLives         = 3
	LifeRefillHours  = 2
	StreakGraceHours = 24
	StreakRecoveryHours = 24
)

type LifeService struct {
	userRepo *repositories.UserRepository
}

func NewLifeService(userRepo *repositories.UserRepository) *LifeService {
	return &LifeService{userRepo: userRepo}
}

func (s *LifeService) RefillLives(user *models.User) {
	if user.LastLifeRefillAt == nil {
		now := time.Now()
		user.LastLifeRefillAt = &now
		return
	}

	elapsed := time.Since(*user.LastLifeRefillAt)
	refillInterval := time.Duration(LifeRefillHours) * time.Hour

	livesToAdd := int(elapsed / refillInterval)
	if livesToAdd <= 0 {
		return
	}

	newLives := user.Lives + livesToAdd
	if newLives > MaxLives {
		newLives = MaxLives
	}

	now := time.Now()
	user.LastLifeRefillAt = &now
	user.Lives = newLives
}

func (s *LifeService) ConsumeLife(user *models.User, lessonID int, progressRepo *repositories.ProgressRepository) {
	if user.Lives > 0 {
		user.Lives--
		return
	}

	// lives == 0 and failed → streak resets + lesson progress deleted
	user.StreakDays = 0
	if lessonID > 0 {
		_ = progressRepo.DeleteByUserAndLesson(user.ID, lessonID)
	}
}

func (s *LifeService) HasLives(user *models.User) bool {
	return user.Lives > 0
}

type LivesStatus struct {
	Current             int   `json:"current"`
	Max                 int   `json:"max"`
	NextRefillInSeconds int64 `json:"next_refill_in_seconds"`
	RefillRateHours     int   `json:"refill_rate_hours"`
}

func (s *LifeService) GetLivesStatus(user *models.User) *LivesStatus {
	nextRefill := int64(0)
	if user.LastLifeRefillAt != nil && user.Lives < MaxLives {
		nextRefillAt := user.LastLifeRefillAt.Add(time.Duration(LifeRefillHours) * time.Hour)
		nextRefill = int64(time.Until(nextRefillAt).Seconds())
		if nextRefill < 0 {
			nextRefill = 0
		}
	}

	return &LivesStatus{
		Current:             user.Lives,
		Max:                 MaxLives,
		NextRefillInSeconds: nextRefill,
		RefillRateHours:     LifeRefillHours,
	}
}
