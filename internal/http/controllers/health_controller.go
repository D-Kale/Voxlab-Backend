package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/database"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthCheck godoc
// @Summary      Health Check
// @Description  Verifies API status, database and Redis connectivity
// @Tags         System
// @Produce      json
// @Success      200  {object}  map[string]interface{}
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
