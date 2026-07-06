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

	exerciseCount := len(lesson.LessonExercises)
	newXP := calculateXP(exerciseCount, input.Score)

	now := time.Now()

	existing, err := s.repo.FindByUserAndLesson(userID, input.LessonID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	diff := newXP
	if existing != nil {
		diff = newXP - existing.XPEarned
		if diff < 0 {
			diff = 0
		}
	}

	progress := &models.UserProgress{
		UserID:      userID,
		LessonID:    input.LessonID,
		Status:      "completed",
		XPEarned:    newXP,
		CompletedAt: &now,
	}

	if err := s.repo.Upsert(progress); err != nil {
		return nil, err
	}

	if diff > 0 {
		_ = s.userRepo.AddXP(userID.String(), diff)
	}

	return progress, nil
}

type UpdateProgressInput struct {
	Score int `json:"score"`
}

func (s *ProgressService) UpdateProgress(userID uuid.UUID, lessonID int, input UpdateProgressInput) (*models.UserProgress, error) {
	lesson, err := s.lessonRepo.FindByID(lessonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("lesson not found")
		}
		return nil, errors.New("lesson not found")
	}

	exerciseCount := len(lesson.LessonExercises)
	newXP := calculateXP(exerciseCount, input.Score)

	existing, err := s.repo.FindByUserAndLesson(userID, lessonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			progress := &models.UserProgress{
				UserID:   userID,
				LessonID: lessonID,
				Status:   "in_progress",
				XPEarned: newXP,
			}
			if err := s.repo.Upsert(progress); err != nil {
				return nil, err
			}
			_ = s.userRepo.AddXP(userID.String(), newXP)
			return progress, nil
		}
		return nil, err
	}

	diff := newXP - existing.XPEarned
	if diff < 0 {
		diff = 0
	}

	existing.XPEarned = newXP
	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	if diff > 0 {
		_ = s.userRepo.AddXP(userID.String(), diff)
	}

	return existing, nil
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
