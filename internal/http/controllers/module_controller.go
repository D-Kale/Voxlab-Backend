package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
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
// @Success      200  {object}  resources.ListModulesResponse   "Módulos del track"
// @Failure      400  {object}  resources.BadRequestError       "ID de track inválido"
// @Failure      500  {object}  resources.InternalServerError   "Error al obtener los módulos"
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
// @Success      200  {object}  resources.GetModuleResponse     "Módulo con lecciones y ejercicios"
// @Failure      400  {object}  resources.BadRequestError       "ID de módulo inválido"
// @Failure      404  {object}  resources.NotFoundError         "Módulo no encontrado"
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
// @Success      201  {object}  resources.CreateModuleResponse  "Módulo creado correctamente"
// @Failure      400  {object}  resources.BadRequestError       "Datos inválidos — título y track_id requeridos"
// @Failure      401  {object}  resources.UnauthorizedError     "Token no proporcionado o inválido"
// @Failure      500  {object}  resources.InternalServerError   "Error al crear el módulo"
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
// @Success      200  {object}  resources.UpdateModuleResponse  "Módulo actualizado correctamente"
// @Failure      400  {object}  resources.BadRequestError       "ID de módulo inválido o datos incorrectos"
// @Failure      401  {object}  resources.UnauthorizedError     "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError         "Módulo no encontrado"
// @Failure      500  {object}  resources.InternalServerError   "Error al actualizar el módulo"
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
// @Success      200  {object}  resources.DeleteModuleResponse  "Módulo eliminado correctamente"
// @Failure      400  {object}  resources.BadRequestError       "ID de módulo inválido"
// @Failure      401  {object}  resources.UnauthorizedError     "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError         "Módulo no encontrado"
// @Failure      500  {object}  resources.InternalServerError   "Error al eliminar el módulo"
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

// ReorderModules godoc
// @Summary      Reorder modules within a track (batch)
// @Description  Updates the order_index for multiple modules at once.
// @Description  Send an array of {id, order_index} pairs.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int   true  "Track ID"
// @Param        request  body  object{items=[]object{id=int,order_index=int}}  true  "Array of module order pairs"
// @Success      200  {object}  map[string]interface{}  "Módulos reordenados"
// @Failure      400  {object}  map[string]interface{}  "Datos inválidos"
// @Failure      401  {object}  map[string]interface{}  "No autorizado"
// @Failure      500  {object}  map[string]interface{}  "Error al reordenar"
// @Router       /api/v1/tracks/{id}/modules/reorder [put]
func (h *ModuleController) ReorderModules(c *gin.Context) {
	var req struct {
		Items []struct {
			ID         int `json:"id"`
			OrderIndex int `json:"order_index"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	items := make([]repositories.ModuleOrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = repositories.ModuleOrderItem{ID: item.ID, OrderIndex: item.OrderIndex}
	}

	if err := h.service.ReorderModules(items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Modules reordered"})
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
// @Success      200  {object}  resources.LinkLessonResponse    "Lección vinculada al módulo correctamente"
// @Failure      400  {object}  resources.BadRequestError       "ID de módulo inválido o datos incorrectos"
// @Failure      401  {object}  resources.UnauthorizedError     "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError         "Módulo o lección no encontrados"
// @Failure      500  {object}  resources.InternalServerError   "Error al vincular la lección"
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

// UnlinkLesson godoc
// @Summary      Unlink a lesson from a module
// @Description  Removes the link between a module and a lesson. The lesson itself is NOT deleted
// @Description  — it remains available for other modules.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Modules
// @Produce      json
// @Security     BearerAuth
// @Param        id        path  int  true  "Module ID (e.g. 1)"
// @Param        lessonId  path  int  true  "Lesson ID (e.g. 1)"
// @Success      200  {object}  resources.UnlinkLessonResponse  "Lección desvinculada del módulo"
// @Failure      400  {object}  resources.BadRequestError        "ID de módulo o lección inválido"
// @Failure      401  {object}  resources.UnauthorizedError      "Token no proporcionado o inválido"
// @Failure      500  {object}  resources.InternalServerError    "Error al desvincular la lección"
// @Router       /api/v1/modules/{id}/lessons/{lessonId} [delete]
func (h *ModuleController) UnlinkLesson(c *gin.Context) {
	moduleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	lessonID, err := strconv.Atoi(c.Param("lessonId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
		return
	}

	if err := h.service.UnlinkLesson(moduleID, lessonID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Lesson unlinked from module"})
}

// ReorderLessons godoc
// @Summary      Reorder lessons within a module (batch)
// @Description  Updates the order_index for multiple lessons in a module.
// @Description  Send an array of {lesson_id, order_index} pairs.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int   true  "Module ID (e.g. 1)"
// @Param        request  body  object{items=[]object{lesson_id=int,order_index=int}}  true  "Array of lesson order pairs"
// @Success      200  {object}  resources.ReorderLessonsResponse  "Lecciones reordenadas"
// @Failure      400  {object}  resources.BadRequestError          "Datos inválidos"
// @Failure      401  {object}  resources.UnauthorizedError        "Token no proporcionado o inválido"
// @Failure      500  {object}  resources.InternalServerError      "Error al reordenar las lecciones"
// @Router       /api/v1/modules/{id}/lessons/reorder [put]
func (h *ModuleController) ReorderLessons(c *gin.Context) {
	moduleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module ID"})
		return
	}

	var req struct {
		Items []struct {
			LessonID   int `json:"lesson_id"`
			OrderIndex int `json:"order_index"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	items := make([]models.ModuleLesson, len(req.Items))
	for i, item := range req.Items {
		items[i] = models.ModuleLesson{ModuleID: moduleID, LessonID: item.LessonID, OrderIndex: item.OrderIndex}
	}

	if err := h.service.ReorderLessons(moduleID, items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Lessons reordered"})
}
