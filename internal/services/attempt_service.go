package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type AttemptService struct {
	lifeSvc      *LifeService
	streakSvc    *StreakService
	exerciseRepo *repositories.ExerciseRepository
	userRepo     *repositories.UserRepository
	attemptRepo  *repositories.ExerciseAttemptRepository
	progressRepo *repositories.ProgressRepository
}

func NewAttemptService(
	lifeSvc *LifeService,
	streakSvc *StreakService,
	exerciseRepo *repositories.ExerciseRepository,
	userRepo *repositories.UserRepository,
	attemptRepo *repositories.ExerciseAttemptRepository,
	progressRepo *repositories.ProgressRepository,
) *AttemptService {
	return &AttemptService{
		lifeSvc:      lifeSvc,
		streakSvc:    streakSvc,
		exerciseRepo: exerciseRepo,
		userRepo:     userRepo,
		attemptRepo:  attemptRepo,
		progressRepo: progressRepo,
	}
}

type AttemptInput struct {
	Score    int `json:"score"`
	LessonID int `json:"lesson_id"`
}

type AttemptResult struct {
	Passed          bool   `json:"passed"`
	Score           int    `json:"score"`
	LivesRemaining  int    `json:"lives_remaining"`
	LivesMax        int    `json:"lives_max"`
	XP              int    `json:"xp"`
	StreakDays      int    `json:"streak_days"`
	StreakAtRisk    bool   `json:"streak_at_risk"`
	StreakLost      bool   `json:"streak_lost"`
	LessonReset     bool   `json:"lesson_reset"`
}

func (s *AttemptService) RegisterAttempt(userID uuid.UUID, exerciseID uuid.UUID, input AttemptInput) (*AttemptResult, error) {
	user, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return nil, errors.New("user not found")
	}

	exercise, err := s.exerciseRepo.FindByID(exerciseID)
	if err != nil {
		return nil, errors.New("exercise not found")
	}

	s.lifeSvc.RefillLives(user)

	passingScore := exercise.PassingScore
	if passingScore <= 0 {
		passingScore = 60
	}

	passed := input.Score >= passingScore
	consumedLife := false
	lessonReset := false

	if !passed {
		if user.Lives == 0 {
			s.lifeSvc.ConsumeLife(user, input.LessonID, s.progressRepo)
			lessonReset = true
		} else {
			user.Lives--
			consumedLife = true
		}
	}

	now := time.Now()
	attempt := &models.ExerciseAttempt{
		ID:           uuid.New(),
		UserID:       userID,
		ExerciseID:   exerciseID,
		LessonID:     input.LessonID,
		Score:        input.Score,
		Passed:       passed,
		ConsumedLife: consumedLife,
		CreatedAt:    now,
	}

	if err := s.attemptRepo.Create(attempt); err != nil {
		return nil, errors.New("failed to register attempt")
	}

	s.streakSvc.TouchActivity(user)
	user.LastActivityAt = &now
	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("failed to update user")
	}

	isAtRisk := s.streakSvc.IsStreakAtRisk(user)
	isLost := s.streakSvc.IsStreakLost(user)

	return &AttemptResult{
		Passed:          passed,
		Score:           input.Score,
		LivesRemaining:  user.Lives,
		LivesMax:        MaxLives,
		XP:              user.XP,
		StreakDays:      user.StreakDays,
		StreakAtRisk:    isAtRisk,
		StreakLost:      isLost,
		LessonReset:     lessonReset,
	}, nil
}
