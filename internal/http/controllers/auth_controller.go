package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type AuthController struct {
	service *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{service: service}
}

// Login godoc
// @Summary      User Login
// @Description  Authenticates with email + password and returns a JWT token.
// @Description  The token must be sent as `Authorization: Bearer <token>` for protected endpoints.
// @Description  Tokens expire after 24 hours.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      services.LoginRequest  true  "Login credentials"
// @Success      200  {object}  resources.LoginResponseData  "Login exitoso — token + datos del usuario"
// @Failure      400  {object}  resources.BadRequestError   "Credenciales inválidas o formato incorrecto"
// @Failure      401  {object}  resources.UnauthorizedError "Email o contraseña incorrectos"
// @Router       /api/v1/auth/login [post]
func (h *AuthController) Login(c *gin.Context) {
	var req services.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	resp, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Register godoc
// @Summary      User Registration
// @Description  Creates a new user account and returns a JWT token (auto-login).
// @Description  The password must be at least 6 characters.
// @Description  If the email is already registered, returns a 409 conflict error.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      services.RegisterRequest  true  "Registration data"
// @Success      201  {object}  resources.RegisterResponseData  "Usuario creado — auto-login con token"
// @Failure      400  {object}  resources.BadRequestError       "Datos de registro inválidos"
// @Failure      409  {object}  resources.ConflictError         "El email ya está registrado"
// @Router       /api/v1/auth/register [post]
func (h *AuthController) Register(c *gin.Context) {
	var req services.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	resp, err := h.service.Register(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "email already registered" {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Logout godoc
// @Summary      User Logout
// @Description  Invalidates the current JWT token by adding it to a Redis blacklist.
// @Description  After calling this, the token can no longer be used for authenticated requests.
// @Description  The frontend should also discard the token locally.
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  resources.LogoutResponse       "Sesión cerrada correctamente"
// @Failure      401  {object}  resources.UnauthorizedError    "Token no proporcionado o inválido"
// @Router       /api/v1/auth/logout [post]
func (h *AuthController) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) < 8 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token not provided"})
		return
	}

	tokenString := authHeader[7:]

	if err := h.service.Logout(tokenString); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "logged out successfully",
	})
}

// Me godoc
// @Summary      Get Current User
// @Description  Returns the authenticated user's profile (name, email, XP, streak, lives).
// @Description  Use this to verify the token is valid and load the user's data on page refresh.
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  resources.UserProfileResponse  "Perfil del usuario autenticado"
// @Failure      401  {object}  resources.UnauthorizedError    "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError        "Usuario no encontrado"
// @Failure      500  {object}  resources.InternalServerError  "Error al obtener el perfil"
// @Router       /api/v1/auth/me [get]
func (h *AuthController) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	user, err := h.service.GetMe(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": services.UserData{
			ID:         user.ID,
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
			AvatarURL:  user.AvatarURL,
			XP:         user.XP,
			StreakDays: user.StreakDays,
		},
	})
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Returns the authenticated user's extended profile (name, email, avatar_url, XP, streak, lives).
// @Description  Unlike /auth/me, this returns a richer profile object with all user-facing fields.
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  resources.UserProfileResponse  "Perfil extendido del usuario"
// @Failure      401  {object}  resources.UnauthorizedError    "Token no proporcionado o inválido"
// @Failure      404  {object}  resources.NotFoundError        "Usuario no encontrado"
// @Failure      500  {object}  resources.InternalServerError  "Error al obtener el perfil"
// @Router       /api/v1/auth/profile [get]
func (h *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	profile, err := h.service.GetProfile(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": profile})
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates the authenticated user's profile fields (name, avatar_url, etc.).
// @Description  Send only the fields you want to change.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      services.UpdateProfileRequest  true  "Profile fields to update"
// @Success      200  {object}  resources.UserProfileResponse  "Perfil actualizado correctamente"
// @Failure      400  {object}  resources.BadRequestError      "Datos inválidos en la solicitud"
// @Failure      401  {object}  resources.UnauthorizedError    "Token no proporcionado o inválido"
// @Failure      500  {object}  resources.InternalServerError  "Error al actualizar el perfil"
// @Router       /api/v1/auth/profile [put]
func (h *AuthController) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	profile, err := h.service.UpdateProfile(userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": profile})
}
