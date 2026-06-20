package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type LessonController struct {
	service *services.LessonService
}

func NewLessonController(service *services.LessonService) *LessonController {
	return &LessonController{service: service}
}

// CreateLessonRequest represents the request body for creating a lesson
// @Description Request body for creating a new lesson
type CreateLessonRequest struct {
	Title               string `json:"title" example:"Respiración Diafragmática"`
	Description         string `json:"description" example:"Aprende a respirar desde el diafragma para proyectar tu voz"`
	EstimatedTimeSeconds int    `json:"estimated_time_seconds" example:"300"`
}

// UpdateLessonRequest represents the request body for updating a lesson
// @Description Request body for updating an existing lesson
type UpdateLessonRequest struct {
	Title                *string `json:"title,omitempty" example:"Respiración Avanzada"`
	Description          *string `json:"description,omitempty" example:"Técnicas avanzadas de control respiratorio"`
	EstimatedTimeSeconds *int    `json:"estimated_time_seconds,omitempty" example:"600"`
}

// GetLessonsByModule godoc
// @Summary      List lessons in a module
// @Description  Returns all lessons linked to a specific module, with their exercises.
// @Description  Lessons are returned through the ModuleLesson pivot and include an order_index.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Lessons
// @Produce      json
// @Param        id   path      int  true  "Module ID (e.g. 1)"
// @Success      200  {object}  resources.ListLessonsResponse   "Lecciones del módulo"
// @Failure      400  {object}  resources.BadRequestError       "ID de módulo inválido"
// @Failure      500  {object}  resources.InternalServerError   "Error al obtener las lecciones"
// @Router       /api/v1/modules/{id}/lessons [get]
func (h *LessonController) GetLessonsByModule(c *gin.Context) {
	moduleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	lessons, err := h.service.GetAllByModule(moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lessons"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    lessons,
	})
}

// GetLesson      godoc
// @Summary      Get a single lesson by ID
// @Description  Returns one lesson with its exercises. Use this to load the full lesson content
// @Description  including all exercise data.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Lessons
// @Produce      json
// @Param        id   path      int  true  "Lesson ID (e.g. 1)"
// @Success      200  {object}  resources.GetLessonResponse     "Lección con ejercicios"
// @Failure      400  {object}  resources.BadRequestError       "ID de lección inválido"
// @Failure      404  {object}  resources.NotFoundError         "Lección no encontrada"
// @Router       /api/v1/lessons/{id} [get]
func (h *LessonController) GetLesson(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	lesson, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    lesson,
	})
}

// CreateLesson   godoc
// @Summary      Create a new lesson
// @Description  Creates a standalone lesson. After creation, link it to a module using
// @Description  POST /api/v1/modules/:id/lessons. Each lesson contains exercises (created separately).
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Lessons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  CreateLessonRequest  true  "Lesson data"
// @Success      201  {object}  resources.CreateLessonResponse  "Lección creada correctamente"
// @Failure      400  {object}  resources.BadRequestError       "Datos inválidos — el título es requerido"
// @Failure      401  {object}  resources.UnauthorizedError     "Token no proporcionado o inválido"
// @Failure      500  {object}  resources.InternalServerError   "Error al crear la lección"
// @Router       /api/v1/lessons [post]
func (h *LessonController) CreateLesson(c *gin.Context) {
	var lesson models.Lesson
	if err := c.ShouldBindJSON(&lesson); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if lesson.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	if err := h.service.Create(&lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lesson"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    lesson,
	})
}

// UpdateLesson   godoc
// @Summary      Update a lesson
// @Description  Changes the title, description, or estimated time of a lesson.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Lessons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path            int                 true  "Lesson ID (e.g. 1)"
// @Param        request  body            UpdateLessonRequest true  "Fields to update"
// @Success      200  {object}  resources.UpdateLessonResponse  "Lección actualizada correctamente"
// @Failure      400  {object}  resources.BadRequestError       "ID de lección inválido o datos incorrectos"
// @Failure      401  {object}  resources.UnauthorizedError     "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError         "Lección no encontrada"
// @Failure      500  {object}  resources.InternalServerError   "Error al actualizar la lección"
// @Router       /api/v1/lessons/{id} [put]
func (h *LessonController) UpdateLesson(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	existing, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
		return
	}

	var input models.Lesson
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if input.Title != "" {
		existing.Title = input.Title
	}
	if input.Description != "" {
		existing.Description = input.Description
	}
	if input.EstimatedTimeSeconds != 0 {
		existing.EstimatedTimeSeconds = input.EstimatedTimeSeconds
	}

	if err := h.service.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lesson"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    existing,
	})
}

// DeleteLesson   godoc
// @Summary      Delete a lesson
// @Description  Permanently removes a lesson and its exercises.
// @Description  ⚠️ This action cannot be undone.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Lessons
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Lesson ID (e.g. 1)"
// @Success      200  {object}  resources.DeleteLessonResponse  "Lección eliminada correctamente"
// @Failure      400  {object}  resources.BadRequestError       "ID de lección inválido"
// @Failure      401  {object}  resources.UnauthorizedError     "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError         "Lección no encontrada"
// @Failure      500  {object}  resources.InternalServerError   "Error al eliminar la lección"
// @Router       /api/v1/lessons/{id} [delete]
func (h *LessonController) DeleteLesson(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	if _, err := h.service.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lesson"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Lesson deleted successfully",
	})
}

// ListLessons godoc
// @Summary      List all lessons globally
// @Description  Returns ALL lessons in the system (global list). Use this to populate
// @Description  search/selection modals when linking lessons to modules.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Lessons
// @Produce      json
// @Success      200  {object}  resources.ListLessonsResponse   "Todas las lecciones"
// @Failure      500  {object}  resources.InternalServerError   "Error al obtener las lecciones"
// @Router       /api/v1/lessons [get]
func (h *LessonController) ListLessons(c *gin.Context) {
	lessons, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lessons"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": lessons})
}

// GetModulesByLesson godoc
// @Summary      Get modules containing a lesson
// @Description  Returns all modules that a specific lesson belongs to, with order_index.
// @Description  Useful for knowing "where is this lesson used?"
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Lessons
// @Produce      json
// @Param        id   path  int  true  "Lesson ID (e.g. 1)"
// @Success      200  {object}  resources.GetModulesByLessonResponse  "Módulos que contienen esta lección"
// @Failure      400  {object}  resources.BadRequestError              "ID de lección inválido"
// @Failure      500  {object}  resources.InternalServerError          "Error al obtener los módulos"
// @Router       /api/v1/lessons/{id}/modules [get]
func (h *LessonController) GetModulesByLesson(c *gin.Context) {
	lessonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	modules, err := h.service.GetModulesByLesson(lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch modules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": modules})
}

// GetExercisesByLesson godoc
// @Summary      Get exercises for a lesson
// @Description  Returns all exercises linked to a specific lesson via the pivot table.
// @Description  Each result includes the exercise data and its order_index from the pivot.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Lessons
// @Produce      json
// @Param        id   path  int  true  "Lesson ID (e.g. 1)"
// @Success      200  {object}  resources.GetExercisesByLessonResponse  "Ejercicios de la lección (con order_index)"
// @Failure      400  {object}  resources.BadRequestError                "ID de lección inválido"
// @Failure      500  {object}  resources.InternalServerError            "Error al obtener los ejercicios"
// @Router       /api/v1/lessons/{id}/exercises [get]
func (h *LessonController) GetExercisesByLesson(c *gin.Context) {
	lessonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	exercises, err := h.service.GetExercisesByLesson(lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exercises"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": exercises})
}
