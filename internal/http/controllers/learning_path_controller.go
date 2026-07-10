package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type LearningPathController struct {
	trackRepo    *repositories.TrackRepository
	moduleRepo   *repositories.ModuleRepository
	lessonRepo   *repositories.LessonRepository
	progressRepo *repositories.ProgressRepository
	lifeSvc      *services.LifeService
	streakSvc    *services.StreakService
}

func NewLearningPathController(
	trackRepo *repositories.TrackRepository,
	moduleRepo *repositories.ModuleRepository,
	lessonRepo *repositories.LessonRepository,
	progressRepo *repositories.ProgressRepository,
	lifeSvc *services.LifeService,
	streakSvc *services.StreakService,
) *LearningPathController {
	return &LearningPathController{
		trackRepo:    trackRepo,
		moduleRepo:   moduleRepo,
		lessonRepo:   lessonRepo,
		progressRepo: progressRepo,
		lifeSvc:      lifeSvc,
		streakSvc:    streakSvc,
	}
}

type LearningPathResponse struct {
	Tracks []TrackWithProgress `json:"tracks"`
}

type TrackWithProgress struct {
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	IconURL     string                 `json:"icon_url"`
	OrderIndex  int                    `json:"order_index"`
	Modules     []ModuleWithProgress   `json:"modules"`
}

type ModuleWithProgress struct {
	ID           int                    `json:"id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	ImageURL     string                 `json:"image_url"`
	OrderIndex   int                    `json:"order_index"`
	Lessons      []LessonWithProgress   `json:"lessons"`
}

type LessonWithProgress struct {
	ID                   int                      `json:"id"`
	Title                string                   `json:"title"`
	Description          string                   `json:"description"`
	EstimatedTimeSeconds int                      `json:"estimated_time_seconds"`
	OrderIndex           int                      `json:"order_index"`
	Status               string                   `json:"status"`
	XPEarned             int                      `json:"xp_earned"`
	CompletedAt          *string                  `json:"completed_at,omitempty"`
	Exercises            []ExerciseWithStatus     `json:"exercises"`
}

type ExerciseWithStatus struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	OrderIndex   int    `json:"order_index"`
	PassingScore int    `json:"passing_score"`
}

// GetLearningPath godoc
// @Summary      Get complete learning path for a track
// @Description  Returns the full tree (modules → lessons → exercises) for a track,
// @Description  with the authenticated user's progress state and locking. All content
// @Description  is ordered by order_index. Lessons within a module are sequentially
// @Description  locked: lesson N is available only when all previous lessons in the
// @Description  same module have status "completed".
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Learning Path
// @Produce      json
// @Param        track_id  query  int  false  "Track ID (returns all tracks if omitted)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Full learning path with progress"
// @Failure      401  {object}  map[string]interface{}  "No autorizado"
// @Failure      500  {object}  map[string]interface{}  "Error interno"
// @Router       /api/v1/learning-path [get]
func (h *LearningPathController) GetLearningPath(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	progressMap := make(map[int]*ProgressInfo)
	allProgress, err := h.progressRepo.FindAllByUser(userID)
	if err == nil {
		for i := range allProgress {
			p := &allProgress[i]
			progressMap[p.LessonID] = &ProgressInfo{
				Status:      p.Status,
				XPEarned:    p.XPEarned,
				CompletedAt: p.CompletedAt,
			}
		}
	}

	trackIDStr := c.Query("track_id")
	var tracks []TrackWithProgress

	if trackIDStr != "" {
		trackID, err := strconv.Atoi(trackIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track_id"})
			return
		}
		track, err := h.trackRepo.FindByID(trackID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Track not found"})
			return
		}
		tracks = []TrackWithProgress{h.buildTrackWithProgress(*track, progressMap)}
	} else {
		allTracks, err := h.trackRepo.FindAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tracks"})
			return
		}
		for _, track := range allTracks {
			tracks = append(tracks, h.buildTrackWithProgress(track, progressMap))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": LearningPathResponse{
			Tracks: tracks,
		},
	})
}

type ProgressInfo struct {
	Status      string
	XPEarned    int
	CompletedAt *time.Time
}

func (h *LearningPathController) buildTrackWithProgress(track models.Track, progressMap map[int]*ProgressInfo) TrackWithProgress {
	tw := TrackWithProgress{
		ID:          track.ID,
		Title:       track.Title,
		Description: track.Description,
		IconURL:     track.IconURL,
		OrderIndex:  track.OrderIndex,
	}

	for _, mod := range track.Modules {
		mw := ModuleWithProgress{
			ID:          mod.ID,
			Title:       mod.Title,
			Description: mod.Description,
			ImageURL:    mod.ImageURL,
			OrderIndex:  mod.OrderIndex,
		}

		allPreviousCompleted := true

		for _, ml := range mod.Lessons {
			lesson := ml.Lesson
			lw := LessonWithProgress{
				ID:                   lesson.ID,
				Title:                lesson.Title,
				Description:          lesson.Description,
				EstimatedTimeSeconds: lesson.EstimatedTimeSeconds,
				OrderIndex:           ml.OrderIndex,
				Status:               "not_started",
			}

			if p, ok := progressMap[lesson.ID]; ok {
				lw.Status = p.Status
				lw.XPEarned = p.XPEarned
				if p.CompletedAt != nil {
					s := p.CompletedAt.Format(time.RFC3339)
					lw.CompletedAt = &s
				}
			}

			if lw.Status == "completed" {
				allPreviousCompleted = true
			} else if allPreviousCompleted {
				allPreviousCompleted = false
			} else {
				lw.Status = "locked"
			}

			for _, le := range lesson.LessonExercises {
				ew := ExerciseWithStatus{
					ID:           le.ExerciseID.String(),
					Name:         le.Exercise.Name,
					Type:         string(le.Exercise.Type),
					OrderIndex:   le.OrderIndex,
					PassingScore: le.Exercise.PassingScore,
				}
				lw.Exercises = append(lw.Exercises, ew)
			}

			mw.Lessons = append(mw.Lessons, lw)
		}

		tw.Modules = append(tw.Modules, mw)
	}

	return tw
}
