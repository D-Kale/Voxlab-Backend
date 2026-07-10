package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type ProgressController struct {
	service *services.ProgressService
}

func NewProgressController(service *services.ProgressService) *ProgressController {
	return &ProgressController{service: service}
}

// GetMyProgress  godoc
// @Summary      Get my learning progress
// @Description  Returns ALL progress records for the authenticated user (every lesson they've started or completed).
// @Description  Each record shows: status (in_progress/completed), xp_earned, and timestamps.
// @Description  Use this on the frontend to determine which lessons are completed and show user progress bars.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Progress
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  resources.ListProgressResponse    "Registros de progreso del usuario"
// @Failure      401  {object}  resources.UnauthorizedError       "Token no proporcionado o inválido"
// @Failure      500  {object}  resources.InternalServerError     "Error al obtener el progreso"
// @Router       /api/v1/progress [get]
func (h *ProgressController) GetMyProgress(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	progress, err := h.service.GetByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    progress,
	})
}

type completeLessonRequest struct {
	LessonID int `json:"lesson_id" example:"1"`
	Score    int `json:"score" example:"85"`
}

// CompleteLesson godoc
// @Summary      Complete a lesson
// @Description  Marks a lesson as completed for the authenticated user. This endpoint:
// @Description  1. Sets the progress status to "completed"
// @Description  2. Adds XP to the user's total (calculated from exercises + score)
// @Description  3. Stores the completion timestamp
// @Description
// @Description  If the lesson was already completed before, it UPDATES the existing record
// @Description  (the final score replaces the previous one).
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Progress
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  completeLessonRequest  true  "Lesson completion data"
// @Success      200  {object}  resources.CompleteProgressResponse  "Progreso actualizado — lección completada"
// @Failure      400  {object}  resources.BadRequestError            "Datos inválidos — lesson_id requerido"
// @Failure      401  {object}  resources.UnauthorizedError          "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError              "Lección no encontrada"
// @Failure      500  {object}  resources.InternalServerError        "Error al completar la lección"
// @Router       /api/v1/progress [post]
func (h *ProgressController) CompleteLesson(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	var req completeLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if req.LessonID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lesson_id is required"})
		return
	}

	progress, err := h.service.CompleteLesson(userID, services.CompleteLessonInput{
		LessonID: req.LessonID,
		Score:    req.Score,
	})
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		if msg == "lesson not found" {
			status = http.StatusNotFound
		} else if msg == "lesson is locked" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    progress,
	})
}

type updateProgressRequest struct {
	Score int `json:"score" example:"85"`
}

// UpdateProgress godoc
// @Summary      Update lesson progress
// @Description  Incrementally updates the XP earned for a lesson without marking it as completed.
// @Description  Calculates the XP diff from the previous value and grants only the delta (never double-counts).
// @Description  If no progress record exists yet, creates one in "in_progress" status.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Progress
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        lesson_id  path  int  true  "Lesson ID"
// @Param        request    body  updateProgressRequest  true  "Updated score"
// @Success      200  {object}  resources.UpdateProgressResponse  "Progreso actualizado"
// @Failure      400  {object}  resources.BadRequestError          "Datos inválidos"
// @Failure      401  {object}  resources.UnauthorizedError        "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError            "Lección no encontrada"
// @Failure      500  {object}  resources.InternalServerError      "Error al actualizar el progreso"
// @Router       /api/v1/progress/{lesson_id} [patch]
func (h *ProgressController) UpdateProgress(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	lessonID, err := strconv.Atoi(c.Param("lesson_id"))
	if err != nil || lessonID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson_id"})
		return
	}

	var req updateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	progress, err := h.service.UpdateProgress(userID, lessonID, services.UpdateProgressInput{
		Score: req.Score,
	})
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		if msg == "lesson not found" {
			status = http.StatusNotFound
		} else if msg == "lesson is locked" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    progress,
	})
}

type syncProgressItem struct {
	LessonID    int        `json:"lesson_id" example:"1"`
	Score       int        `json:"score" example:"85"`
	Status      string     `json:"status" example:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type syncProgressRequest struct {
	ProgressItems []syncProgressItem `json:"progress_items"`
}

// SyncProgress godoc
// @Summary      Bulk sync local progress (first-time registration only)
// @Description  Syncs ALL locally-stored lesson progress to the server after registration.
// @Description  Only applies if the user has ZERO XP and ZERO streak days (first-time sync).
// @Description  If the user already has server-side progress, returns 409 Conflict.
// @Description  The frontend groups exercises by lesson and sends lesson-level scores.
// @Description  After sync, the frontend should re-fetch GET /progress and GET /auth/me
// @Description  to replace local state with server state.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Progress
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  syncProgressRequest  true  "List of lesson progress items to sync"
// @Success      200  {object}  resources.SyncProgressResponse  "Progreso sincronizado — incluye progress[] y user actualizado"
// @Failure      400  {object}  resources.BadRequestError        "Datos inválidos"
// @Failure      401  {object}  resources.UnauthorizedError      "Token no proporcionado o inválido"
// @Failure      409  {object}  resources.ConflictError          "El usuario ya tiene progreso en el servidor"
// @Failure      500  {object}  resources.InternalServerError    "Error al sincronizar el progreso"
// @Router       /api/v1/progress/sync [post]
func (h *ProgressController) SyncProgress(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	var req syncProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if len(req.ProgressItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "progress_items is required"})
		return
	}

	svcItems := make([]services.SyncProgressItem, len(req.ProgressItems))
	for i, item := range req.ProgressItems {
		svcItems[i] = services.SyncProgressItem{
			LessonID:    item.LessonID,
			Score:       item.Score,
			Status:      item.Status,
			CompletedAt: item.CompletedAt,
		}
	}

	progress, user, err := h.service.SyncProgress(userID, services.SyncProgressInput{
		ProgressItems: svcItems,
	})
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		if msg == "user already has progress data, sync rejected" {
			status = http.StatusConflict
		} else if msg == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"progress": progress,
			"user":     user,
		},
	})
}
