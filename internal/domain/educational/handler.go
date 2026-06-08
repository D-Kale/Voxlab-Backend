package educational

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetTracks(c *gin.Context) {
	tracks, err := h.service.GetTracks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo tracks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tracks,
	})
}
