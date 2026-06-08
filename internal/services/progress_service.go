package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
	"gorm.io/gorm"
)

type ProgressService struct {
	repo      *repositories.ProgressRepository
	lessonRepo *repositories.LessonRepository
	userRepo  *repositories.UserRepository
}

func NewProgressService(
	repo *repositories.ProgressRepository,
	lessonRepo *repositories.LessonRepository,
	userRepo *repositories.UserRepository,
) *ProgressService {
	return &ProgressService{
		repo:       repo,
		lessonRepo: lessonRepo,
		userRepo:   userRepo,
	}
}

type CompleteLessonInput struct {
	LessonID int  `json:"lesson_id"`
	Score    int  `json:"score"`
}

func (s *ProgressService) CompleteLesson(userID uuid.UUID, input CompleteLessonInput) (*models.UserProgress, error) {
	lesson, err := s.lessonRepo.FindByID(input.LessonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("lesson not found")
		}
		return nil, errors.New("lesson not found")
	}

	exerciseCount := len(lesson.Exercises)
	xpEarned := calculateXP(exerciseCount, input.Score)

	now := time.Now()
	progress := &models.UserProgress{
		UserID:      userID,
		LessonID:    input.LessonID,
		Status:      "completed",
		XPEarned:    xpEarned,
		CompletedAt: &now,
	}

	if err := s.repo.Upsert(progress); err != nil {
		return nil, err
	}

	_ = s.userRepo.AddXP(userID.String(), xpEarned)

	return progress, nil
}

func calculateXP(exerciseCount, score int) int {
	baseXP := 10
	if exerciseCount > 0 {
		baseXP += exerciseCount * 5
	}
	if score > 0 {
		baseXP += score / 10
	}
	return baseXP
}

func (s *ProgressService) GetByUser(userID uuid.UUID) ([]models.UserProgress, error) {
	return s.repo.FindAllByUser(userID)
}

func (s *ProgressService) GetByUserAndLesson(userID uuid.UUID, lessonID int) (*models.UserProgress, error) {
	return s.repo.FindByUserAndLesson(userID, lessonID)
}
