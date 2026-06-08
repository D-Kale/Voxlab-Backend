package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/config"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.MustGetConfig()
		allowOrigin := "*"
		if cfg.AppEnv == "production" {
			allowOrigin = "https://voxlab.app"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
