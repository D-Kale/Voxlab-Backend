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

// CreateReaction godoc
// @Summary      [TODO] Create a reaction (emoji/like) during live sessions
// @Description  ⚠️ NOT YET IMPLEMENTED — placeholder endpoint.
// @Description  This will allow users to send reactions (emojis/likes) during
// @Description  live practice sessions. Reactions are stored with a 30-day TTL
// @Description  and automatically cleaned up by pg_cron.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Reactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      501  {object}  resources.NotImplementedResponse  "Funcionalidad no implementada aún"
// @Router       /api/v1/reactions [post]
func (h *ReactionController) CreateReaction(c *gin.Context) {
	c.JSON(501, gin.H{"message": "not implemented"})
}
