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
// @Description  Authenticates credentials and returns a JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      services.LoginRequest  true  "Login credentials"
// @Success      200     {object}  services.LoginResponse
// @Failure      400     {object}  map[string]string
// @Failure      401     {object}  map[string]string
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
