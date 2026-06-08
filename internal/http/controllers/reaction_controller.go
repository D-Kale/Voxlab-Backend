package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type ReactionController struct {
	service *services.CommunityService
}

func NewReactionController(service *services.CommunityService) *ReactionController {
	return &ReactionController{service: service}
}

func (h *ReactionController) CreateReaction(c *gin.Context) {
	c.JSON(501, gin.H{"message": "not implemented"})
}
