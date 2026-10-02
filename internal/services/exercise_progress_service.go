package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
)

type ExerciseProgressCompleteInput struct {
	ExerciseID  string     `json:"exercise_id" binding:"required"`
	Score       int        `json:"score" binding:"required,min=0,max=100"`
	CompletedAt time.Time  `json:"completed_at" binding:"required"`
	LessonID    *int       `json:"lesson_id,omitempty"`
}

type ExerciseProgressSyncInput struct {
	Entries []ExerciseProgressCompleteInput `json:"entries" binding:"required,min=1"`
}

type ExerciseProgressResult struct {
	XP              int      `json:"xp"`
	StreakDays      int      `json:"streak_days"`
	NewlyCompleted  []string `json:"newly_completed"`
	StreakUpdated   bool     `json:"streak_updated"`
}

type UserProgressResponse struct {
	XP                   int                   `json:"xp"`
	StreakDays           int                   `json:"streak_days"`
	CompletedExerciseIds []string              `json:"completed_exercise_ids"`
	WeeklyChart          []WeeklyChartPoint    `json:"weekly_chart"`
	LastSyncedAt         *time.Time            `json:"last_synced_at,omitempty"`
}

type WeeklyChartPoint = repositories.WeeklyChartPoint

type ExerciseProgressService struct {
	repo          *repositories.ExerciseProgressRepository
	userRepo      *repositories.UserRepository
	exerciseRepo  *repositories.ExerciseRepository
}

func NewExerciseProgressService(
	repo *repositories.ExerciseProgressRepository,
	userRepo *repositories.UserRepository,
	exerciseRepo *repositories.ExerciseRepository,
) *ExerciseProgressService {
	return &ExerciseProgressService{
		repo:          repo,
		userRepo:      userRepo,
		exerciseRepo:  exerciseRepo,
	}
}

func (s *ExerciseProgressService) CompleteExercise(userID uuid.UUID, input ExerciseProgressCompleteInput) (*ExerciseProgressResult, error) {
	// 1. Verify exercise exists
	_, err := s.exerciseRepo.FindByID(uuid.MustParse(input.ExerciseID))
	if err != nil {
		return nil, errors.New("exercise not found")
	}

	// 2. Validate completedAt is not too far in future (5 min tolerance)
	if input.CompletedAt.After(time.Now().Add(5 * time.Minute)) {
		return nil, errors.New("completed_at cannot be in the future")
	}

	// 3. Check if already completed (idempotency)
	existing, err := s.repo.FindByUserAndExercise(userID, input.ExerciseID)
	if err == nil && existing != nil {
		// Already completed - return current state, no error
		return s.getProgressResponse(userID), nil
	}

	// 4. Get user
	user, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 5. Calculate XP gained (score direct, no bonus)
	xpGained := input.Score

	// 6. Calculate streak delta
	streakDelta := calculateStreakDelta(user.LastCompletedAt, input.CompletedAt)
	newStreak := user.StreakDays + streakDelta
	if newStreak < 1 {
		newStreak = 1
	}
	streakUpdated := streakDelta != 0

	// 7. Update user
	user.XP += xpGained
	user.StreakDays = newStreak
	user.LastCompletedAt = &input.CompletedAt
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// 8. Insert progress record
	progress := &models.UserExerciseProgress{
		UserID:      userID,
		ExerciseID:  input.ExerciseID,
		LessonID:    input.LessonID,
		Score:       input.Score,
		CompletedAt: input.CompletedAt,
	}
	if err := s.repo.UpsertExerciseProgress(progress); err != nil {
		return nil, err
	}

	// 9. Return result
	return &ExerciseProgressResult{
		XP:             user.XP,
		StreakDays:     user.StreakDays,
		NewlyCompleted: []string{input.ExerciseID},
		StreakUpdated:  streakUpdated,
	}, nil
}

func (s *ExerciseProgressService) SyncProgress(userID uuid.UUID, input ExerciseProgressSyncInput) (*ExerciseProgressResult, error) {
	// Sort entries by completedAt ascending
	for i := 0; i < len(input.Entries)-1; i++ {
		for j := i + 1; j < len(input.Entries); j++ {
			if input.Entries[i].CompletedAt.After(input.Entries[j].CompletedAt) {
				input.Entries[i], input.Entries[j] = input.Entries[j], input.Entries[i]
			}
		}
	}

	user, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Reset user state for recalculation
	user.XP = 0
	user.StreakDays = 0
	user.LastCompletedAt = nil

	var newlyCompleted []string
	var lastCompletedAt *time.Time

	for _, entry := range input.Entries {
		// Validate exercise
		_, err := s.exerciseRepo.FindByID(uuid.MustParse(entry.ExerciseID))
		if err != nil {
			continue
		}

		// Validate timestamp
		if entry.CompletedAt.After(time.Now().Add(5 * time.Minute)) {
			continue
		}

		// Check if already exists
		existing, _ := s.repo.FindByUserAndExercise(userID, entry.ExerciseID)
		if existing != nil {
			// Update if score is higher
			if entry.Score > existing.Score {
				existing.Score = entry.Score
				existing.CompletedAt = entry.CompletedAt
				if entry.LessonID != nil {
					existing.LessonID = entry.LessonID
				}
				_ = s.repo.UpsertExerciseProgress(existing)
			}
			continue
		}

		// New entry
		xpGained := entry.Score
		user.XP += xpGained

		streakDelta := calculateStreakDelta(lastCompletedAt, entry.CompletedAt)
		user.StreakDays += streakDelta
		if user.StreakDays < 1 {
			user.StreakDays = 1
		}

		progress := &models.UserExerciseProgress{
			UserID:      userID,
			ExerciseID:  entry.ExerciseID,
			LessonID:    entry.LessonID,
			Score:       entry.Score,
			CompletedAt: entry.CompletedAt,
		}
		if err := s.repo.UpsertExerciseProgress(progress); err != nil {
			continue
		}

		newlyCompleted = append(newlyCompleted, entry.ExerciseID)
		lastCompletedAt = &entry.CompletedAt
	}

	if user.StreakDays < 1 {
		user.StreakDays = 1
	}
	user.LastCompletedAt = lastCompletedAt
	_ = s.userRepo.Update(user)

	streakUpdated := len(newlyCompleted) > 0
	return &ExerciseProgressResult{
		XP:             user.XP,
		StreakDays:     user.StreakDays,
		NewlyCompleted: newlyCompleted,
		StreakUpdated:  streakUpdated,
	}, nil
}

func (s *ExerciseProgressService) GetProgress(userID uuid.UUID) (*UserProgressResponse, error) {
	user, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return nil, errors.New("user not found")
	}

	completedExercises, _ := s.repo.FindCompletedExercises(userID)
	weeklyChart, _ := s.repo.GetWeeklyChart(userID, 14)

	var lastSyncedAt *time.Time
	if !user.UpdatedAt.IsZero() {
		lastSyncedAt = &user.UpdatedAt
	}

	return &UserProgressResponse{
		XP:                   user.XP,
		StreakDays:           user.StreakDays,
		CompletedExerciseIds: completedExercises,
		WeeklyChart:          weeklyChart,
		LastSyncedAt:         lastSyncedAt,
	}, nil
}

func (s *ExerciseProgressService) getProgressResponse(userID uuid.UUID) *ExerciseProgressResult {
	user, _ := s.userRepo.FindByID(userID.String())
	return &ExerciseProgressResult{
		XP:             user.XP,
		StreakDays:     user.StreakDays,
		NewlyCompleted: []string{},
		StreakUpdated:  false,
	}
}

func calculateStreakDelta(lastCompletedAt *time.Time, currentCompletedAt time.Time) int {
	if lastCompletedAt == nil {
		return 1 // First exercise ever
	}

	lastDate := lastCompletedAt.Truncate(24 * time.Hour)
	currentDate := currentCompletedAt.Truncate(24 * time.Hour)

	daysDiff := int(currentDate.Sub(lastDate).Hours() / 24)

	if daysDiff == 0 {
		return 0 // Same day
	}
	if daysDiff == 1 {
		return 1 // Consecutive day
	}
	// Gap > 1 day: reset to 1 (new streak starts today)
	return 1 - 0 // Returns 1, streak will be set to 1 in calling code
}