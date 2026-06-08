package controllers

import (
	"net/http"

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
// @Success      200  {object}  map[string]interface{}  "Success: { success: true, data: UserProgress[] }"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized — token missing or invalid"
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
// @Success      200  {object}  map[string]interface{}  "Completed: { success: true, data: UserProgress }"
// @Failure      400  {object}  map[string]interface{}  "Validation error"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized — token missing or invalid"
// @Failure      404  {object}  map[string]interface{}  "Lesson not found"
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
		if err.Error() == "lesson not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    progress,
	})
}
