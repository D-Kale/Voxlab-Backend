package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	attemptRepo *repositories.ExerciseAttemptRepository
}

func NewProgressService(
	repo *repositories.ProgressRepository,
	lessonRepo *repositories.LessonRepository,
	userRepo *repositories.UserRepository,
	lifeSvc *LifeService,
	streakSvc *StreakService,
	attemptRepo *repositories.ExerciseAttemptRepository,
) *ProgressService {
	return &ProgressService{
		repo:        repo,
		lessonRepo:  lessonRepo,
		userRepo:    userRepo,
		lifeSvc:     lifeSvc,
		streakSvc:   streakSvc,
		attemptRepo: attemptRepo,
	}
}

type CompleteExerciseInput struct {
	ExerciseID string `json:"exercise_id"`
	Score      int    `json:"score"`
}

type CompleteLessonInput struct {
	LessonID  int                    `json:"lesson_id"`
	Score     int                    `json:"score"`
	Exercises []CompleteExerciseInput `json:"exercises"`
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

	// Obtener ejercicios de la lección para validar exercise_ids
	lessonExercises, err := s.lessonRepo.FindExercisesByLesson(input.LessonID)
	if err != nil {
		return nil, err
	}

	validExerciseIDs := make(map[string]bool)
	for _, le := range lessonExercises {
		validExerciseIDs[le.ExerciseID.String()] = true
	}

	now := time.Now()

	existing, err := s.repo.FindByUserAndLesson(userID, input.LessonID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Parsear ejercicios completados existentes
	var existingExercises []models.CompletedExercise
	if existing != nil && existing.CompletedExercises != nil {
		_ = json.Unmarshal(existing.CompletedExercises, &existingExercises)
		if existingExercises == nil {
			existingExercises = []models.CompletedExercise{}
		}
	}

	// Procesar ejercicios nuevos, validando duplicados
	newExercises := make([]models.CompletedExercise, 0, len(input.Exercises))
	newXP := 0

	for _, exInput := range input.Exercises {
		// Validar que el exercise_id pertenece a la lección
		if !validExerciseIDs[exInput.ExerciseID] {
			return nil, fmt.Errorf("exercise %s not in lesson %d", exInput.ExerciseID, input.LessonID)
		}

		// Verificar si ya está completado
		alreadyCompleted := false
		for _, ex := range existingExercises {
			if ex.ExerciseID == exInput.ExerciseID {
				alreadyCompleted = true
				break
			}
		}

		if !alreadyCompleted {
			newExercises = append(newExercises, models.CompletedExercise{
				ExerciseID:  exInput.ExerciseID,
				Score:       exInput.Score,
				CompletedAt: &now,
			})
			newXP += scoreToXP(exInput.Score)
		}
	}

	// Merge: ejercicios existentes + nuevos
	allExercises := append(existingExercises, newExercises...)
	completedExercisesJSON, err := json.Marshal(allExercises)
	if err != nil {
		return nil, err
	}

	accumulated := newXP
	if existing != nil {
		accumulated = existing.XPEarned + newXP
	}

	progress := &models.UserProgress{
		UserID:             userID,
		LessonID:           input.LessonID,
		Status:             "completed",
		XPEarned:           accumulated,
		CompletedExercises: completedExercisesJSON,
		CompletedAt:        &now,
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
	Score     int                    `json:"score"`
	Exercises []CompleteExerciseInput `json:"exercises"`
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

	// Obtener ejercicios de la lección para validar
	lessonExercises, err := s.lessonRepo.FindExercisesByLesson(lessonID)
	if err != nil {
		return nil, err
	}

	validExerciseIDs := make(map[string]bool)
	for _, le := range lessonExercises {
		validExerciseIDs[le.ExerciseID.String()] = true
	}

	newXP := 0
	newExercises := make([]models.CompletedExercise, 0, len(input.Exercises))
	now := time.Now()

	existing, err := s.repo.FindByUserAndLesson(userID, lessonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No existe progreso previo, crear uno nuevo
			for _, exInput := range input.Exercises {
				if !validExerciseIDs[exInput.ExerciseID] {
					return nil, fmt.Errorf("exercise %s not in lesson %d", exInput.ExerciseID, lessonID)
				}
				newExercises = append(newExercises, models.CompletedExercise{
					ExerciseID:  exInput.ExerciseID,
					Score:       exInput.Score,
					CompletedAt: &now,
				})
				newXP += scoreToXP(exInput.Score)
			}

			completedExercisesJSON, err := json.Marshal(newExercises)
			if err != nil {
				return nil, err
			}

			progress := &models.UserProgress{
				UserID:             userID,
				LessonID:           lessonID,
				Status:             "in_progress",
				XPEarned:           newXP,
				CompletedExercises: completedExercisesJSON,
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
		return nil, err
	}

	if existing.Status == "completed" {
		return existing, nil
	}

	// Parsear ejercicios existentes
	var existingExercises []models.CompletedExercise
	if existing.CompletedExercises != nil {
		_ = json.Unmarshal(existing.CompletedExercises, &existingExercises)
		if existingExercises == nil {
			existingExercises = []models.CompletedExercise{}
		}
	}

	// Procesar nuevos ejercicios, validando duplicados
	for _, exInput := range input.Exercises {
		if !validExerciseIDs[exInput.ExerciseID] {
			return nil, fmt.Errorf("exercise %s not in lesson %d", exInput.ExerciseID, lessonID)
		}

		alreadyCompleted := false
		for _, ex := range existingExercises {
			if ex.ExerciseID == exInput.ExerciseID {
				alreadyCompleted = true
				break
			}
		}

		if !alreadyCompleted {
			newExercises = append(newExercises, models.CompletedExercise{
				ExerciseID:  exInput.ExerciseID,
				Score:       exInput.Score,
				CompletedAt: &now,
			})
			newXP += scoreToXP(exInput.Score)
		}
	}

	// Merge y serialización
	allExercises := append(existingExercises, newExercises...)
	completedExercisesJSON, err := json.Marshal(allExercises)
	if err != nil {
		return nil, err
	}

	existing.XPEarned += newXP
	existing.CompletedExercises = completedExercisesJSON
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

type DailyCount struct {
	Weekday int `json:"weekday"`
	Count   int `json:"count"`
}

type WeeklyProgressData struct {
	WeekStart string       `json:"week_start"`
	Daily     []DailyCount `json:"daily"`
}

func (s *ProgressService) GetWeeklyProgress(userID uuid.UUID) (*WeeklyProgressData, error) {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	weekStart := now.AddDate(0, 0, -int(weekday-time.Monday))
	weekEnd := weekStart.AddDate(0, 0, 7)

	dailyCounts := make([]int, 7)

	// 1. Count from exercise_attempts
	attempts, err := s.attemptRepo.CountPassedAttemptsThisWeek(userID)
	if err != nil {
		return nil, err
	}
	for _, a := range attempts {
		w := a.Date.Weekday()
		idx := int(w) - 1
		if idx < 0 {
			idx = 6
		}
		dailyCounts[idx] += a.Count
	}

	// 2. Count from JSONB completed_exercises (for exercises saved without exercise_attempts)
	progressList, err := s.repo.FindAllByUser(userID)
	if err != nil {
		return nil, err
	}
	for _, p := range progressList {
		var exercises []models.CompletedExercise
		if p.CompletedExercises != nil {
			_ = json.Unmarshal(p.CompletedExercises, &exercises)
		}
		for _, ex := range exercises {
			if ex.CompletedAt != nil && !ex.CompletedAt.Before(weekStart) && ex.CompletedAt.Before(weekEnd) {
				w := ex.CompletedAt.Weekday()
				idx := int(w) - 1
				if idx < 0 {
					idx = 6
				}
				dailyCounts[idx]++
			}
		}
	}

	daily := make([]DailyCount, 7)
	for i := 0; i < 7; i++ {
		daily[i] = DailyCount{
			Weekday: i,
			Count:   dailyCounts[i],
		}
	}

	return &WeeklyProgressData{
		WeekStart: weekStart.Format("2006-01-02"),
		Daily:     daily,
	}, nil
}
