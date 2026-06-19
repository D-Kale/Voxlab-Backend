package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/database"
)

// HealthResponse represents the system health status
// @Description System health check response
type HealthResponse struct {
	Status    string            `json:"status" example:"ok"`
	Timestamp string            `json:"timestamp" example:"2026-06-18T12:00:00Z"`
	Version   string            `json:"version" example:"1.0.0"`
	Services  map[string]string `json:"services"` // e.g. {"postgres": "ok", "redis": "ok"}
}

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthCheck godoc
// @Summary      Health Check
// @Description  Verifies API status, database and Redis connectivity.
// @Description  Returns current version and service-level health for each dependency.
// @Tags         System
// @Produce      json
// @Success      200  {object}  controllers.HealthResponse  "Estado del sistema — ok"
// @Failure      503  {object}  resources.ServiceUnavailableError  "Servicio no saludable — postgres o redis caídos"
// @Router       /api/v1/health [get]
func (h *HealthController) HealthCheck(c *gin.Context) {
	db := database.GetDB()
	rdb := database.GetRedis()

	dbStatus := "ok"
	if sqlDB, err := db.DB(); err != nil {
		dbStatus = "error: " + err.Error()
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = "error: " + err.Error()
	}

	redisStatus := "ok"
	if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
		redisStatus = "error: " + err.Error()
	}

	c.JSON(200, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"services": gin.H{
			"postgres": dbStatus,
			"redis":    redisStatus,
		},
	})
}
