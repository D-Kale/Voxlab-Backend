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
	repo        *repositories.ProgressRepository
	lessonRepo  *repositories.LessonRepository
	userRepo    *repositories.UserRepository
	lifeSvc     *LifeService
	streakSvc   *StreakService
}

func NewProgressService(
	repo *repositories.ProgressRepository,
	lessonRepo *repositories.LessonRepository,
	userRepo *repositories.UserRepository,
	lifeSvc *LifeService,
	streakSvc *StreakService,
) *ProgressService {
	return &ProgressService{
		repo:       repo,
		lessonRepo: lessonRepo,
		userRepo:   userRepo,
		lifeSvc:    lifeSvc,
		streakSvc:  streakSvc,
	}
}

type CompleteLessonInput struct {
	LessonID int `json:"lesson_id"`
	Score    int `json:"score"`
}

func (s *ProgressService) CompleteLesson(userID uuid.UUID, input CompleteLessonInput) (*models.UserProgress, error) {
	_, err := s.lessonRepo.FindByID(input.LessonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("lesson not found")
		}
		return nil, errors.New("lesson not found")
	}

	locked, err := s.CheckLessonLocked(userID, input.LessonID)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, errors.New("lesson is locked")
	}

	newXP := scoreToXP(input.Score)

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

	s.touchUserActivity(userID)

	return progress, nil
}

type UpdateProgressInput struct {
	Score int `json:"score"`
}

func (s *ProgressService) UpdateProgress(userID uuid.UUID, lessonID int, input UpdateProgressInput) (*models.UserProgress, error) {
	_, err := s.lessonRepo.FindByID(lessonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("lesson not found")
		}
		return nil, errors.New("lesson not found")
	}

	locked, err := s.CheckLessonLocked(userID, lessonID)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, errors.New("lesson is locked")
	}

	newXP := scoreToXP(input.Score)

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
			s.touchUserActivity(userID)
			return progress, nil
		}
		return nil, err
	}

	if existing.Status == "completed" {
		return existing, nil
	}

	existing.XPEarned += newXP
	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	if newXP > 0 {
		_ = s.userRepo.AddXP(userID.String(), newXP)
		s.grantDailyStreak(userID)
	}

	s.touchUserActivity(userID)

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
		_, err := s.lessonRepo.FindByID(item.LessonID)
		if err != nil {
			continue
		}

		xpEarned := scoreToXP(item.Score)

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
		s.touchUserActivity(userID)
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

	if s.streakSvc.IsStreakAtRisk(user) {
		if user.Lives > 0 {
			user.Lives--
		} else {
			return
		}
	}

	user.StreakDays++
	if err := s.userRepo.Update(user); err != nil {
		return
	}

	_ = database.SetUserStreak(ctx, userID.String())
	_ = database.TrackUserProgress(ctx, userID.String(), user.XP, user.StreakDays)
}

func (s *ProgressService) touchUserActivity(userID uuid.UUID) {
	user, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return
	}
	s.streakSvc.TouchActivity(user)
	_ = s.userRepo.Update(user)
}

func scoreToXP(score int) int {
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

func (s *ProgressService) CheckLessonLocked(userID uuid.UUID, lessonID int) (bool, error) {
	moduleLessonLinks, err := s.lessonRepo.FindModulesByLesson(lessonID)
	if err != nil || len(moduleLessonLinks) == 0 {
		return false, errors.New("lesson not found in any module")
	}

	moduleID := moduleLessonLinks[0].ModuleID

	moduleLessons, err := s.lessonRepo.FindAllByModule(moduleID)
	if err != nil {
		return false, err
	}

	var lessonIDs []int
	for _, ml := range moduleLessons {
		lessonIDs = append(lessonIDs, ml.LessonID)
	}

	progressList, err := s.repo.FindByUserAndLessons(userID, lessonIDs)
	if err != nil {
		return false, err
	}

	progressMap := make(map[int]string)
	for _, p := range progressList {
		progressMap[p.LessonID] = p.Status
	}

	for _, ml := range moduleLessons {
		if ml.LessonID == lessonID {
			return false, nil
		}
		status, exists := progressMap[ml.LessonID]
		if !exists || status != "completed" {
			return true, nil
		}
	}

	return false, nil
}
