package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type ModuleController struct {
	service *services.ModuleService
}

func NewModuleController(service *services.ModuleService) *ModuleController {
	return &ModuleController{service: service}
}

// CreateModuleRequest represents the request body for creating a module
// @Description Request body for creating a new module
type CreateModuleRequest struct {
	TrackID     int    `json:"track_id" example:"1"`
	Title       string `json:"title" example:"Voz y Proyección"`
	Description string `json:"description" example:"Técnicas para proyectar la voz sin esfuerzo"`
	OrderIndex  int    `json:"order_index" example:"1"`
}

// UpdateModuleRequest represents the request body for updating a module
// @Description Request body for updating an existing module
type UpdateModuleRequest struct {
	Title       *string `json:"title,omitempty" example:"Voz y Dicción"`
	Description *string `json:"description,omitempty" example:"Técnicas avanzadas de vocalización"`
	OrderIndex  *int    `json:"order_index,omitempty" example:"2"`
}

type linkLessonRequest struct {
	LessonID int `json:"lesson_id" example:"1"`
}

// GetModulesByTrack godoc
// @Summary      List modules for a track
// @Description  Returns all modules belonging to a specific track, ordered by order_index.
// @Description  Each module includes its linked lessons and their exercises.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Modules
// @Produce      json
// @Param        id   path      int  true  "Track ID (e.g. 1)"
// @Success      200  {array}   models.Module  "List of modules"
// @Failure      400  {object}  map[string]interface{}  "Invalid track ID"
// @Failure      500  {object}  map[string]interface{}  "Server error"
// @Router       /api/v1/tracks/{id}/modules [get]
func (h *ModuleController) GetModulesByTrack(c *gin.Context) {
	trackID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	modules, err := h.service.GetAllByTrack(trackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch modules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    modules,
	})
}

// GetModule      godoc
// @Summary      Get a single module by ID
// @Description  Returns one module with its linked lessons and exercises.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Modules
// @Produce      json
// @Param        id   path      int  true  "Module ID (e.g. 1)"
// @Success      200  {object}  models.Module  "Module details"
// @Failure      400  {object}  map[string]interface{}  "Invalid module ID"
// @Failure      404  {object}  map[string]interface{}  "Module not found"
// @Router       /api/v1/modules/{id} [get]
func (h *ModuleController) GetModule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	module, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    module,
	})
}

// CreateModule   godoc
// @Summary      Create a new module
// @Description  Adds a module inside a specific track (course).
// @Description  The track_id must reference an existing track. Modules appear in order_index order.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  CreateModuleRequest  true  "Module data"
// @Success      201  {object}  models.Module  "Created module"
// @Failure      400  {object}  map[string]interface{}  "Validation error"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Router       /api/v1/modules [post]
func (h *ModuleController) CreateModule(c *gin.Context) {
	var module models.Module
	if err := c.ShouldBindJSON(&module); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if module.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	if err := h.service.Create(&module); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create module"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    module,
	})
}

// UpdateModule   godoc
// @Summary      Update a module
// @Description  Changes the title, description, or order of a module.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path    int                 true  "Module ID (e.g. 1)"
// @Param        request  body    UpdateModuleRequest true  "Fields to update"
// @Success      200  {object}  models.Module  "Updated module"
// @Failure      400  {object}  map[string]interface{}  "Validation error"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      404  {object}  map[string]interface{}  "Module not found"
// @Router       /api/v1/modules/{id} [put]
func (h *ModuleController) UpdateModule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	existing, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	var input models.Module
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
	if input.OrderIndex != 0 {
		existing.OrderIndex = input.OrderIndex
	}

	if err := h.service.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update module"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    existing,
	})
}

// DeleteModule   godoc
// @Summary      Delete a module
// @Description  Permanently removes a module and its lesson links. Lessons themselves are NOT deleted,
// @Description  only the link between the module and the lesson is removed.
// @Description  ⚠️ This action cannot be undone.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Modules
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Module ID (e.g. 1)"
// @Success      200  {object}  map[string]interface{}  "Deleted: { success: true, message: string }"
// @Failure      400  {object}  map[string]interface{}  "Invalid module ID"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      404  {object}  map[string]interface{}  "Module not found"
// @Failure      500  {object}  map[string]interface{}  "Server error"
// @Router       /api/v1/modules/{id} [delete]
func (h *ModuleController) DeleteModule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	if _, err := h.service.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete module"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Module deleted successfully",
	})
}

// LinkLesson     godoc
// @Summary      Link a lesson to a module
// @Description  Associates an existing lesson with a module using the ModuleLesson pivot table.
// @Description  A lesson can be linked to MULTIPLE modules. This does NOT move or copy the lesson.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int  true  "Module ID (e.g. 1)"
// @Param        request  body  linkLessonRequest  true  "Lesson ID to link"
// @Success      200  {object}  map[string]interface{}  "Linked: { success: true, message: string }"
// @Failure      400  {object}  map[string]interface{}  "Validation error"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      404  {object}  map[string]interface{}  "Module or Lesson not found"
// @Router       /api/v1/modules/{id}/lessons [post]
func (h *ModuleController) LinkLesson(c *gin.Context) {
	moduleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	var req linkLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := h.service.LinkLesson(moduleID, req.LessonID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Lesson linked to module successfully",
	})
}
