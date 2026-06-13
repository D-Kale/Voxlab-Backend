package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type ExerciseController struct {
	service *services.ExerciseService
}

func NewExerciseController(service *services.ExerciseService) *ExerciseController {
	return &ExerciseController{service: service}
}

// GetExercisesByLesson godoc
// @Summary      List exercises for a lesson
// @Description  Returns all exercises in a lesson, ordered by order_index.
// @Description  Each exercise has a "type" field that defines the JSON structure of its "content" field.
// @Description  See the "content" field descriptions below for each exercise type.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Exercises
// @Produce      json
// @Param        id   path      int  true  "Lesson ID (e.g. 1)"
// @Success      200  {object}  map[string]interface{}  "Success: { success: true, data: Exercise[] }"
// @Failure      404  {object}  map[string]interface{}  "Lesson not found"
// @Router       /api/v1/lessons/{id}/exercises [get]
func (h *ExerciseController) GetExercisesByLesson(c *gin.Context) {
	lessonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	exercises, err := h.service.GetAllByLesson(lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exercises"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    exercises,
	})
}

// GetExercise    godoc
// @Summary      Get a single exercise by ID
// @Description  Returns one exercise with its full JSONB content. The content structure depends
// @Description  on the exercise type (quiz, reading, oratory_minigame, audio, video, writing).
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Exercises
// @Produce      json
// @Param        id   path      string  true  "Exercise UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)"
// @Success      200  {object}  map[string]interface{}  "Success: { success: true, data: Exercise }"
// @Failure      404  {object}  map[string]interface{}  "Exercise not found"
// @Router       /api/v1/exercises/{id} [get]
func (h *ExerciseController) GetExercise(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID (must be a valid UUID)"})
		return
	}

	exercise, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    exercise,
	})
}

// CreateExercise godoc
// @Summary      Create a new exercise
// @Description  Creates an exercise inside a lesson. The "type" field determines the JSONB "content" structure.
// @Description
// @Description  📝 Supported exercise types and their content structure:
// @Description
// @Description  **quiz** — Multiple choice questions (multi-pregunta):
// @Description  ```json
// @Description  {
// @Description    "type": "quiz",
// @Description    "content": {
// @Description      "questions": [
// @Description        {
// @Description          "question": "What is public speaking?",
// @Description          "options": ["Option A", "Option B", "Option C", "Option D"],
// @Description          "correct_index": 0,
// @Description          "explanation": "Option A is correct because..."
// @Description        }
// @Description      ],
// @Description      "points_per_question": 10
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **reading** — Reading passage:
// @Description  ```json
// @Description  {
// @Description    "type": "reading",
// @Description    "content": {
// @Description      "title": "The Art of Speech",
// @Description      "content": "Full reading text here...",
// @Description      "reading_time_seconds": 120,
// @Description      "points": 5
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **oratory_minigame** — Oratory challenge with requirements:
// @Description  ```json
// @Description  {
// @Description    "type": "oratory_minigame",
// @Description    "content": {
// @Description      "prompt": "Record a 30-second speech about...",
// @Description      "topic": "Leadership",
// @Description      "duration_seconds": 30,
// @Description      "min_duration_seconds": 15,
// @Description      "requirements": ["Clear introduction", "Use at least 3 key points", "Strong conclusion"],
// @Description      "points": 20
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **writing** — Writing exercise with requirements:
// @Description  ```json
// @Description  {
// @Description    "type": "writing",
// @Description    "content": {
// @Description      "prompt": "Write a 200-word essay about leadership",
// @Description      "min_words": 100,
// @Description      "max_words": 500,
// @Description      "requirements": ["Include a thesis", "Support with examples"],
// @Description      "points": 20
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **audio** — Audio recording exercise:
// @Description  ```json
// @Description  {
// @Description    "type": "audio",
// @Description    "content": {
// @Description      "prompt": "Read this paragraph aloud...",
// @Description      "duration_seconds": 60,
// @Description      "points": 15
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **video** — Video recording exercise:
// @Description  ```json
// @Description  {
// @Description    "type": "video",
// @Description    "content": {
// @Description      "prompt": "Record a video introducing yourself...",
// @Description      "duration_seconds": 120,
// @Description      "points": 25
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Exercises
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  object{lesson_id=int,type=string,order_index=int,content=object}  true  "Exercise data"
// @Success      201  {object}  map[string]interface{}  "Created: { success: true, data: Exercise }"
// @Failure      400  {object}  map[string]interface{}  "Validation error"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Router       /api/v1/exercises [post]
func (h *ExerciseController) CreateExercise(c *gin.Context) {
	var exercise models.Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if exercise.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type is required. Accepted: quiz, reading, oratory_minigame, audio, video, writing"})
		return
	}

	if exercise.LessonID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lesson_id is required"})
		return
	}

	exercise.ID = uuid.New()

	if err := h.service.Create(&exercise); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exercise"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    exercise,
	})
}

// UpdateExercise godoc
// @Summary      Update an exercise
// @Description  Modifies the type, content (JSONB), or order of an exercise.
// @Description  When updating the content field, send the FULL new content object for the exercise type.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Exercises
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string  true  "Exercise UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)"
// @Param        request  body  object{type=string,order_index=int,content=object}  true  "Fields to update"
// @Success      200  {object}  map[string]interface{}  "Updated: { success: true, data: Exercise }"
// @Failure      400  {object}  map[string]interface{}  "Validation error"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      404  {object}  map[string]interface{}  "Exercise not found"
// @Router       /api/v1/exercises/{id} [put]
type updateExerciseInput struct {
	Type       models.ExerciseType `json:"type,omitempty"`
	Content    json.RawMessage     `json:"content,omitempty"`
	OrderIndex *int                `json:"order_index,omitempty"`
}

func (h *ExerciseController) UpdateExercise(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID (must be a valid UUID)"})
		return
	}

	existing, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
		return
	}

	var input updateExerciseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if input.Type != "" {
		existing.Type = input.Type
	}
	if input.Content != nil {
		existing.Content = input.Content
	}
	if input.OrderIndex != nil {
		existing.OrderIndex = *input.OrderIndex
	}

	if err := h.service.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exercise"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    existing,
	})
}

// DeleteExercise godoc
// @Summary      Delete an exercise
// @Description  Permanently removes an exercise from a lesson.
// @Description  ⚠️ This action cannot be undone.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Exercises
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Exercise UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)"
// @Success      200  {object}  map[string]interface{}  "Deleted: { success: true, message: string }"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      404  {object}  map[string]interface{}  "Exercise not found"
// @Router       /api/v1/exercises/{id} [delete]
func (h *ExerciseController) DeleteExercise(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID (must be a valid UUID)"})
		return
	}

	if _, err := h.service.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete exercise"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Exercise deleted successfully",
	})
}
