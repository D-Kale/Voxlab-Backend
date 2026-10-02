package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type ExerciseProgressController struct {
	service *services.ExerciseProgressService
}

func NewExerciseProgressController(service *services.ExerciseProgressService) *ExerciseProgressController {
	return &ExerciseProgressController{service: service}
}

type completeExerciseRequest struct {
	ExerciseID  string    `json:"exercise_id" binding:"required"`
	Score       int       `json:"score" binding:"required,min=0,max=100"`
	CompletedAt time.Time `json:"completed_at" binding:"required"`
	LessonID    *int      `json:"lesson_id,omitempty"`
}

type syncExerciseProgressRequest struct {
	Entries []completeExerciseRequest `json:"entries" binding:"required,min=1"`
}

// GetProgress godoc
// @Summary      Get user exercise progress
// @Description  Returns complete progress state for the authenticated user.
// @Description  Includes total XP, streak, completed exercise IDs, weekly chart.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Progress (Exercise-based)
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Estado completo de progreso"
// @Failure      401  {object}  map[string]interface{}  "No autorizado"
// @Router       /api/v1/progress [get]
func (h *ExerciseProgressController) GetProgress(c *gin.Context) {
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

	response, err := h.service.GetProgress(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

// CompleteExercise godoc
// @Summary      Complete an exercise
// @Description  Records an exercise completion for the authenticated user.
// @Description  Idempotent: if already completed, returns current state without error.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Progress (Exercise-based)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  completeExerciseRequest  true  "Exercise completion data"
// @Success      200  {object}  map[string]interface{}  "Progreso actualizado"
// @Failure      400  {object}  map[string]interface{}  "Datos inválidos"
// @Failure      401  {object}  map[string]interface{}  "No autorizado"
// @Failure      404  {object}  map[string]interface{}  "Ejercicio no encontrado"
// @Router       /api/v1/progress [post]
func (h *ExerciseProgressController) CompleteExercise(c *gin.Context) {
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

	var req completeExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	result, err := h.service.CompleteExercise(userID, services.ExerciseProgressCompleteInput{
		ExerciseID:  req.ExerciseID,
		Score:       req.Score,
		CompletedAt: req.CompletedAt,
		LessonID:    req.LessonID,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "exercise not found" {
			status = http.StatusNotFound
		} else if err.Error() == "invalid exercise_id format (must be UUID)" || err.Error() == "completed_at cannot be in the future" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// SyncProgress godoc
// @Summary      Bulk sync progress (migration)
// @Description  Syncs multiple exercise completions at once (first-time migration).
// @Description  Recalculates XP and streak from all entries ordered by completedAt.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Progress (Exercise-based)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  syncProgressRequest  true  "List of exercise completions"
// @Success      200  {object}  map[string]interface{}  "Estado consolidado tras sync"
// @Failure      400  {object}  map[string]interface{}  "Datos inválidos"
// @Failure      401  {object}  map[string]interface{}  "No autorizado"
// @Router       /api/v1/progress/sync [post]
func (h *ExerciseProgressController) SyncProgress(c *gin.Context) {
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

	var req syncExerciseProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	svcEntries := make([]services.ExerciseProgressCompleteInput, len(req.Entries))
	for i, e := range req.Entries {
		svcEntries[i] = services.ExerciseProgressCompleteInput{
			ExerciseID:  e.ExerciseID,
			Score:       e.Score,
			CompletedAt: e.CompletedAt,
			LessonID:    e.LessonID,
		}
	}

	result, err := h.service.SyncProgress(userID, services.ExerciseProgressSyncInput{
		Entries: svcEntries,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}