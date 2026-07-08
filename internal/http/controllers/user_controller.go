package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/services"
	"gorm.io/gorm"
)

type UserController struct {
	service   *services.UserService
	lifeSvc   *services.LifeService
	streakSvc *services.StreakService
	db        *gorm.DB
}

func NewUserController(service *services.UserService, lifeSvc *services.LifeService, streakSvc *services.StreakService, db *gorm.DB) *UserController {
	return &UserController{service: service, lifeSvc: lifeSvc, streakSvc: streakSvc, db: db}
}

func (h *UserController) getFullUser(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := h.db.First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, false
	}
	return userID, true
}

// GetUsers godoc
// @Summary      List all users (admin)
// @Description  Returns all registered users. Admin only. Includes profile data (name, email, XP, streak).
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  resources.ListUsersResponse     "Lista de todos los usuarios"
// @Failure      401  {object}  resources.UnauthorizedError     "Token no proporcionado o inválido"
// @Failure      403  {object}  resources.ForbiddenError        "Solo administradores pueden listar usuarios"
// @Failure      500  {object}  resources.InternalServerError   "Error al obtener los usuarios"
// @Router       /api/v1/users [get]
func (h *UserController) GetUsers(c *gin.Context) {
	users, err := h.service.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": users})
}

// GetUser godoc
// @Summary      Get user by ID (admin)
// @Description  Returns a single user's profile data. Admin only.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  resources.GetUserResponse      "Datos del usuario"
// @Failure      401  {object}  resources.UnauthorizedError    "Token no proporcionado o inválido"
// @Failure      403  {object}  resources.ForbiddenError       "Solo administradores pueden ver usuarios"
// @Failure      404  {object}  resources.NotFoundError        "Usuario no encontrado"
// @Router       /api/v1/users/{id} [get]
func (h *UserController) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": user})
}

// UpdateUser godoc
// @Summary      Update user (admin)
// @Description  Modifies a user's role, name, email, or other fields. Admin only.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string                      true  "User ID (UUID)"
// @Param        request  body  services.UpdateUserRequest   true  "Fields to update"
// @Success      200  {object}  resources.UpdateUserResponse   "Usuario actualizado correctamente"
// @Failure      400  {object}  resources.BadRequestError      "Datos inválidos en la solicitud"
// @Failure      401  {object}  resources.UnauthorizedError    "Token no proporcionado o inválido"
// @Failure      403  {object}  resources.ForbiddenError       "Solo administradores pueden modificar usuarios"
// @Failure      404  {object}  resources.NotFoundError        "Usuario no encontrado"
// @Failure      500  {object}  resources.InternalServerError  "Error al actualizar el usuario"
// @Router       /api/v1/users/{id} [put]
func (h *UserController) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	user, err := h.service.UpdateUser(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": user})
}

// DeleteUser godoc
// @Summary      Delete user (admin)
// @Description  Permanently removes a user account. Admin only. ⚠️ This cannot be undone.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  resources.DeleteUserResponse   "Usuario eliminado correctamente"
// @Failure      401  {object}  resources.UnauthorizedError    "Token no proporcionado o inválido"
// @Failure      403  {object}  resources.ForbiddenError       "Solo administradores pueden eliminar usuarios"
// @Failure      404  {object}  resources.NotFoundError        "Usuario no encontrado"
// @Failure      500  {object}  resources.InternalServerError  "Error al eliminar el usuario"
// @Router       /api/v1/users/{id} [delete]
func (h *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "user deleted"})
}

// GetLives godoc
// @Summary      Get current lives status
// @Description  Returns the user's current lives, max lives, and next refill time.
// @Description  Lives regenerate 1 every 2 hours up to a maximum of 3.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Estado de vidas"
// @Failure      401  {object}  map[string]interface{}  "No autorizado"
// @Failure      404  {object}  map[string]interface{}  "Usuario no encontrado"
// @Router       /api/v1/users/lives [get]
func (h *UserController) GetLives(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.getFullUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	h.lifeSvc.RefillLives(user)
	_ = h.db.Save(user).Error

	c.JSON(http.StatusOK, gin.H{"success": true, "data": h.lifeSvc.GetLivesStatus(user)})
}

// RecoverStreak godoc
// @Summary      Recover streak by spending a life
// @Description  If the user's streak is at risk (24h without activity), spend 1 life to keep it alive.
// @Description  Can only recover within the 24h grace window after the risk period.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Racha recuperada"
// @Failure      400  {object}  map[string]interface{}  "No se puede recuperar — sin vidas o racha no está en riesgo"
// @Failure      401  {object}  map[string]interface{}  "No autorizado"
// @Failure      404  {object}  map[string]interface{}  "Usuario no encontrado"
// @Router       /api/v1/users/streak/recover [post]
func (h *UserController) RecoverStreak(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.getFullUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	h.lifeSvc.RefillLives(user)

	if !h.streakSvc.IsStreakAtRisk(user) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streak is not at risk, no recovery needed"})
		return
	}

	if err := h.streakSvc.RecoverStreak(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"streak_days": user.StreakDays,
			"lives":       user.Lives,
		},
	})
}
