package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/analyzer"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type ExerciseController struct {
	service    *services.ExerciseService
	attemptSvc *services.AttemptService
}

// CreateExerciseRequest represents the request body for creating an exercise
// @Description Request body for creating an exercise
type CreateExerciseRequest struct {
	// Exercise name (helps identify when reusing across lessons)
	Name string `json:"name" example:"Quiz de liderazgo"`
	// Exercise type: quiz, reading, oratory_minigame, audio, video, writing
	Type models.ExerciseType `json:"type" example:"quiz"`
	// Exercise content (JSONB structure varies by type)
	Content interface{} `json:"content" swaggertype:"object"`
}

func NewExerciseController(service *services.ExerciseService, attemptSvc *services.AttemptService) *ExerciseController {
	return &ExerciseController{service: service, attemptSvc: attemptSvc}
}

// ListExercises godoc
// @Summary      List all exercises
// @Description  Returns all exercises in the system (global list). Exercises are reusable
// @Description  across lessons — use GET /lessons/:id to see which exercises belong to a lesson.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Exercises
// @Produce      json
// @Success      200  {object}  resources.ListExercisesResponse   "Todos los ejercicios"
// @Failure      500  {object}  resources.InternalServerError     "Error al obtener los ejercicios"
// @Router       /api/v1/exercises [get]
func (h *ExerciseController) ListExercises(c *gin.Context) {
	exercises, err := h.service.GetAll()
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
// @Success      200  {object}  resources.GetExerciseResponse  "Detalles del ejercicio"
// @Failure      400  {object}  resources.BadRequestError      "UUID de ejercicio inválido"
// @Failure      404  {object}  resources.NotFoundError        "Ejercicio no encontrado"
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
// @Description  Creates an exercise that can later be linked to one or more lessons.
// @Description  The "type" field determines the JSONB "content" structure.
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
// @Param        request  body  CreateExerciseRequest true  "Exercise data"
// @Success      201  {object}  resources.CreateExerciseResponse  "Ejercicio creado correctamente"
// @Failure      400  {object}  resources.BadRequestError          "Datos inválidos — tipo requerido"
// @Failure      401  {object}  resources.UnauthorizedError        "Token no proporcionado o inválido"
// @Router       /api/v1/exercises [post]
func (h *ExerciseController) CreateExercise(c *gin.Context) {
	var req CreateExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type is required. Accepted: quiz, reading, oratory_minigame, audio, video, writing"})
		return
	}

	content, err := json.Marshal(req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content JSON"})
		return
	}

	exercise := models.Exercise{
		ID:      uuid.New(),
		Name:    req.Name,
		Type:    req.Type,
		Content: content,
	}

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
// @Description  Modifies the name, type, or content (JSONB) of an exercise.
// @Description  When updating the content field, send the FULL new content object for the exercise type.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Exercises
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string  true  "Exercise UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)"
// @Param        request  body  object{name=string,type=string,content=object}  true  "Fields to update"
// @Success      200  {object}  resources.UpdateExerciseResponse  "Ejercicio actualizado correctamente"
// @Failure      400  {object}  resources.BadRequestError          "UUID inválido o datos incorrectos"
// @Failure      401  {object}  resources.UnauthorizedError        "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError            "Ejercicio no encontrado"
// @Router       /api/v1/exercises/{id} [put]
type updateExerciseInput struct {
	Name    *string             `json:"name,omitempty"`
	Type    *models.ExerciseType `json:"type,omitempty"`
	Content json.RawMessage     `json:"content,omitempty"`
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

	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Type != nil {
		existing.Type = *input.Type
	}
	if input.Content != nil {
		existing.Content = input.Content
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
// @Description  Permanently removes an exercise. This also removes all links to lessons.
// @Description  ⚠️ This action cannot be undone.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Exercises
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Exercise UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)"
// @Success      200  {object}  resources.DeleteExerciseResponse  "Ejercicio eliminado correctamente"
// @Failure      400  {object}  resources.BadRequestError          "UUID de ejercicio inválido"
// @Failure      401  {object}  resources.UnauthorizedError        "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError            "Ejercicio no encontrado"
// @Failure      500  {object}  resources.InternalServerError      "Error al eliminar el ejercicio"
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

// ============================================================================
// Lesson ⇄ Exercise links
// ============================================================================

// LinkExerciseToLesson godoc
// @Summary      Link an exercise to a lesson
// @Description  Adds an existing exercise to a lesson via the pivot table.
// @Description  The exercise is appended at the end of the lesson's exercise order.
// @Description  The same exercise can be linked to multiple lessons.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Lessons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        lessonId  path  int     true  "Lesson ID (e.g. 1)"
// @Param        request   body  object{exercise_id=string}  true  "Exercise UUID to link"
// @Success      200  {object}  resources.LinkExerciseResponse  "Ejercicio vinculado a la lección"
// @Failure      400  {object}  resources.BadRequestError       "Datos inválidos"
// @Failure      401  {object}  resources.UnauthorizedError     "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError         "Lección o ejercicio no encontrado"
// @Router       /api/v1/lessons/{id}/exercises [post]
func (h *ExerciseController) LinkExerciseToLesson(c *gin.Context) {
	lessonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	var req struct {
		ExerciseID string `json:"exercise_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	exerciseID, err := uuid.Parse(req.ExerciseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID (must be a valid UUID)"})
		return
	}

	// Verify exercise exists
	if _, err := h.service.GetByID(exerciseID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
		return
	}

	if err := h.service.LinkExerciseToLesson(lessonID, exerciseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link exercise"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Exercise linked to lesson"})
}

// UnlinkExerciseFromLesson godoc
// @Summary      Unlink an exercise from a lesson
// @Description  Removes the link between an exercise and a lesson. The exercise itself is NOT deleted
// @Description  — it remains available for other lessons.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Lessons
// @Produce      json
// @Security     BearerAuth
// @Param        lessonId    path  int     true  "Lesson ID (e.g. 1)"
// @Param        exerciseId  path  string  true  "Exercise UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)"
// @Success      200  {object}  resources.UnlinkExerciseResponse  "Ejercicio desvinculado de la lección"
// @Failure      400  {object}  resources.BadRequestError          "Datos inválidos"
// @Failure      401  {object}  resources.UnauthorizedError        "Token no proporcionado o inválido"
// @Failure      500  {object}  resources.InternalServerError      "Error al desvincular el ejercicio"
// @Router       /api/v1/lessons/{id}/exercises/{exerciseId} [delete]
func (h *ExerciseController) UnlinkExerciseFromLesson(c *gin.Context) {
	lessonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	exerciseID, err := uuid.Parse(c.Param("exerciseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID (must be a valid UUID)"})
		return
	}

	if err := h.service.UnlinkExerciseFromLesson(lessonID, exerciseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink exercise"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Exercise unlinked from lesson"})
}

// ReorderExerciseInLesson godoc
// @Summary      Reorder an exercise within a lesson
// @Description  Updates the display order of an exercise inside a lesson.
// @Description  Lower order_index values appear first.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Lessons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        lessonId    path  int     true  "Lesson ID (e.g. 1)"
// @Param        exerciseId  path  string  true  "Exercise UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)"
// @Param        request     body  object{order_index=int}  true  "New order position"
// @Success      200  {object}  resources.ReorderExerciseResponse  "Orden actualizado"
// @Failure      400  {object}  resources.BadRequestError           "Datos inválidos"
// @Failure      401  {object}  resources.UnauthorizedError         "Token no proporcionado o inválido"
// @Failure      500  {object}  resources.InternalServerError       "Error al reordenar el ejercicio"
// @Router       /api/v1/lessons/{id}/exercises/{exerciseId}/reorder [put]
func (h *ExerciseController) ReorderExerciseInLesson(c *gin.Context) {
	lessonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	exerciseID, err := uuid.Parse(c.Param("exerciseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID (must be a valid UUID)"})
		return
	}

	var req struct {
		OrderIndex int `json:"order_index" example:"1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := h.service.ReorderExerciseInLesson(lessonID, exerciseID, req.OrderIndex); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reorder exercise"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Exercise reordered"})
}

// ============================================================================
// Exercise ⇄ Lesson reverse lookup + batch reorder
// ============================================================================

// GetLessonsByExercise godoc
// @Summary      Get lessons containing an exercise
// @Description  Returns all lessons that a specific exercise belongs to, with order_index.
// @Description  Useful for knowing "where is this exercise used?"
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Exercises
// @Produce      json
// @Param        id   path  string  true  "Exercise UUID (e.g. 550e8400-e29b-41d4-a716-446655440000)"
// @Success      200  {object}  resources.GetLessonsByExerciseResponse  "Lecciones que contienen este ejercicio"
// @Failure      400  {object}  resources.BadRequestError                 "UUID inválido"
// @Failure      500  {object}  resources.InternalServerError             "Error al obtener las lecciones"
// @Router       /api/v1/exercises/{id}/lessons [get]
func (h *ExerciseController) GetLessonsByExercise(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID (must be a valid UUID)"})
		return
	}

	links, err := h.service.GetLessonsByExercise(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lessons"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": links})
}

// BatchReorderExercises godoc
// @Summary      Reorder exercises within a lesson (batch)
// @Description  Updates the order_index for multiple exercises in a lesson at once.
// @Description  Send an array of {exercise_id, order_index} pairs.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Lessons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int   true  "Lesson ID (e.g. 1)"
// @Param        request  body  object{items=[]object{exercise_id=string,order_index=int}}  true  "Array of exercise order pairs"
// @Success      200  {object}  resources.ReorderExercisesResponse  "Ejercicios reordenados"
// @Failure      400  {object}  resources.BadRequestError            "Datos inválidos"
// @Failure      401  {object}  resources.UnauthorizedError          "Token no proporcionado o inválido"
// @Failure      500  {object}  resources.InternalServerError        "Error al reordenar los ejercicios"
// @Router       /api/v1/lessons/{id}/exercises/reorder [put]
func (h *ExerciseController) BatchReorderExercises(c *gin.Context) {
	lessonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	var req struct {
		Items []struct {
			ExerciseID string `json:"exercise_id"`
			OrderIndex int    `json:"order_index"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	items := make([]models.LessonExercise, len(req.Items))
	for i, item := range req.Items {
		exID, err := uuid.Parse(item.ExerciseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise UUID in item " + strconv.Itoa(i)})
			return
		}
		items[i] = models.LessonExercise{LessonID: lessonID, ExerciseID: exID, OrderIndex: item.OrderIndex}
	}

	if err := h.service.BatchReorderExercises(lessonID, items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Exercises reordered"})
}

// ============================================================================
// Exercise Attempt (lives + streak)
// ============================================================================

type attemptExerciseRequest struct {
	Score    int `json:"score" example:"75"`
	LessonID int `json:"lesson_id" example:"1"`
}

// AttemptExercise godoc
// @Summary      Register an exercise attempt
// @Description  Records a user's attempt at an exercise. If the score is below the passing threshold,
// @Description  a life is consumed. If lives reach 0, the streak resets and the current lesson progress is deleted.
// @Description  Lives regenerate over time (1 every 2 hours, max 3).
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Exercises
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string                true  "Exercise UUID"
// @Param        request  body  attemptExerciseRequest true  "Attempt data"
// @Success      200  {object}  map[string]interface{}  "Resultado del intento"
// @Failure      400  {object}  map[string]interface{}  "Datos inválidos"
// @Failure      401  {object}  map[string]interface{}  "No autorizado"
// @Failure      404  {object}  map[string]interface{}  "Ejercicio no encontrado"
// @Router       /api/v1/exercises/{id}/attempt [post]
func (h *ExerciseController) AttemptExercise(c *gin.Context) {
	exerciseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID (must be a valid UUID)"})
		return
	}

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

	var req attemptExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if req.LessonID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lesson_id is required"})
		return
	}

	result, err := h.attemptSvc.RegisterAttempt(userID, exerciseID, services.AttemptInput{
		Score:    req.Score,
		LessonID: req.LessonID,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "exercise not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// ============================================================================
// Text analysis (unchanged)
// ============================================================================

type AnalyzeTextInput struct {
	Text         string   `json:"text" example:"El liderazgo es una habilidad fundamental para cualquier profesional que busque destacar en su carrera. A lo largo de este ensayo, exploraremos las características clave de un líder efectivo y cómo desarrollarlas en el día a día."`
	Requirements []string `json:"requirements,omitempty" example:"Incluir una introducción clara,Dar ejemplos concretos,Usar vocabulario técnico"`
	MinWords     *int     `json:"min_words,omitempty" example:"100"`
	MaxWords     *int     `json:"max_words,omitempty" example:"500"`
}

type RequirementCatalogItem struct {
	ID       string `json:"id" example:"intro"`
	Text     string `json:"text" example:"Incluir una introducción clara del tema"`
	Category string `json:"category" example:"Cobertura del contenido"`
	Order    int    `json:"order" example:"1"`
}

// GetRequirementCatalog godoc
// @Summary      Get requirement catalog for writing exercises
// @Description  Returns the curated list of selectable requirements for writing exercises,
// @Description  grouped by category. The frontend renders these as checkboxes so teachers
// @Description  can pick predefined requirements instead of typing free text.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Exercises
// @Produce      json
// @Success      200  {object}  resources.RequirementCatalogResponse  "Catálogo de requisitos agrupados por categoría"
// @Router       /api/v1/exercises/requirement-catalog [get]
func (h *ExerciseController) GetRequirementCatalog(c *gin.Context) {
	catalog := models.GetRequirementCatalog()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    catalog,
	})
}

// AnalyzeText godoc
// @Summary      Analyze text for writing exercises
// @Description  Sends text to the Python analyzer service for NLP analysis including:
// @Description  gibberish detection, sentence structure, vocabulary richness, readability,
// @Description  semantic requirement matching, and weighted score calculation.
// @Description
// @Description  **Ejemplo de request:**
// @Description  ```json
// @Description  {
// @Description    "text": "El liderazgo es una habilidad fundamental...",
// @Description    "requirements": ["Incluir una introducción", "Dar ejemplos concretos"],
// @Description    "min_words": 100,
// @Description    "max_words": 500
// @Description  }
// @Description  ```
// @Description
// @Description  **Respuesta:** `score` (0-100), `score_breakdown` con cada componente
// @Description  (cobertura_requisitos 30%, estructura 25%, calidad_linguistica 25%,
// @Description  longitud 20%), `gibberish_detected`, métricas y retroalimentación.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Exercises
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  AnalyzeTextInput  true  "Text to analyze"
// @Success      200  {object}  resources.AnalyzeTextResponse      "Análisis completo del texto"
// @Failure      400  {object}  resources.BadRequestError           "Texto vacío o datos inválidos"
// @Failure      401  {object}  resources.UnauthorizedError         "Token no proporcionado o inválido"
// @Failure      502  {object}  resources.ServiceUnavailableError   "Analizador NLP no disponible — intente más tarde"
// @Router       /api/v1/exercises/analyze-text [post]
func (h *ExerciseController) AnalyzeText(c *gin.Context) {
	var input AnalyzeTextInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if input.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
		return
	}

	result, err := analyzer.AnalyzeText(input.Text, input.Requirements, input.MinWords, input.MaxWords)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Analyzer service unavailable", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
