package community

import "github.com/gin-gonic/gin"

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) CreateReaction(c *gin.Context) {
	c.JSON(501, gin.H{"message": "not implemented"})
}
