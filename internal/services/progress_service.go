package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/database"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
	"gorm.io/gorm"
)

type ProgressService struct {
	repo       *repositories.ProgressRepository
	lessonRepo *repositories.LessonRepository
	userRepo   *repositories.UserRepository
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
	LessonID int `json:"lesson_id"`
	Score    int `json:"score"`
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

	accumulated := newXP
	if existing != nil {
		if existing.Status == "completed" {
			return existing, nil
		}
		accumulated = existing.XPEarned + newXP
	}

	progress := &models.UserProgress{
		UserID:      userID,
		LessonID:    input.LessonID,
		Status:      "completed",
		XPEarned:    accumulated,
		CompletedAt: &now,
	}

	if err := s.repo.Upsert(progress); err != nil {
		return nil, err
	}

	if newXP > 0 {
		_ = s.userRepo.AddXP(userID.String(), newXP)
		s.grantDailyStreak(userID)
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
			s.grantDailyStreak(userID)
			return progress, nil
		}
		return nil, err
	}

	if existing.Status == "completed" {
		return existing, nil
	}

	existing.Status = "completed"
	existing.XPEarned += newXP
	existing.CompletedAt = func() *time.Time { t := time.Now(); return &t }()
	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	if newXP > 0 {
		_ = s.userRepo.AddXP(userID.String(), newXP)
		s.grantDailyStreak(userID)
	}

	return existing, nil
}

type SyncProgressItem struct {
	LessonID    int        `json:"lesson_id"`
	Score       int        `json:"score"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type SyncProgressInput struct {
	ProgressItems []SyncProgressItem `json:"progress_items"`
}

func (s *ProgressService) SyncProgress(userID uuid.UUID, input SyncProgressInput) ([]models.UserProgress, *models.User, error) {
	user, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return nil, nil, errors.New("user not found")
	}

	if user.XP > 0 || user.StreakDays > 0 {
		return nil, nil, errors.New("user already has progress data, sync rejected")
	}

	var totalXP int
	var synced []models.UserProgress

	for _, item := range input.ProgressItems {
		lesson, err := s.lessonRepo.FindByID(item.LessonID)
		if err != nil {
			continue
		}

		exerciseCount := len(lesson.LessonExercises)
		xpEarned := calculateXP(exerciseCount, item.Score)

		progress := &models.UserProgress{
			UserID:      userID,
			LessonID:    item.LessonID,
			Status:      item.Status,
			XPEarned:    xpEarned,
			CompletedAt: item.CompletedAt,
		}
		if progress.CompletedAt == nil {
			now := time.Now()
			progress.CompletedAt = &now
		}

		if err := s.repo.Upsert(progress); err != nil {
			continue
		}

		totalXP += xpEarned
		synced = append(synced, *progress)
	}

	user.XP = totalXP
	if err := s.userRepo.Update(user); err != nil {
		return nil, nil, err
	}

	if len(synced) > 0 {
		s.grantDailyStreak(userID)
	}

	updatedUser, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return synced, user, nil
	}

	return synced, updatedUser, nil
}

func (s *ProgressService) grantDailyStreak(userID uuid.UUID) {
	ctx := context.Background()
	hasStreak, err := database.GetUserStreak(ctx, userID.String())
	if err != nil || hasStreak {
		return
	}

	user, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return
	}

	user.StreakDays++
	if err := s.userRepo.Update(user); err != nil {
		return
	}

	_ = database.SetUserStreak(ctx, userID.String())
	_ = database.TrackUserProgress(ctx, userID.String(), user.XP, user.StreakDays)
}

func calculateXP(_ int, score int) int {
	if score < 0 {
		return 0
	}
	return score
}

func (s *ProgressService) GetByUser(userID uuid.UUID) ([]models.UserProgress, error) {
	return s.repo.FindAllByUser(userID)
}

func (s *ProgressService) GetByUserAndLesson(userID uuid.UUID, lessonID int) (*models.UserProgress, error) {
	return s.repo.FindByUserAndLesson(userID, lessonID)
}
