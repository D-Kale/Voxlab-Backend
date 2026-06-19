package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type UserController struct {
	service *services.UserService
}

func NewUserController(service *services.UserService) *UserController {
	return &UserController{service: service}
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
