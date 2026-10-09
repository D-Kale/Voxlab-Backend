package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/logger"
)

// RequestLogger emite una línea de access log por request.
//
// Antes el engine usaba gin.New() sin ningún logger de acceso: no había forma
// de saber qué endpoint se llamó, con qué status ni cuánto tardó. Este middleware
// registra todo eso vía slog, así hereda el formato JSON de production, e
// incluye request_id y user_id para correlacionar con el navegador.
//
// Se registra DESPUÉS de c.Next() para poder capturar el status final y los
// errores acumulados en c.Errors. Por eso debe ir después de RequestID().
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fullPath := path
		if query != "" {
			fullPath = path + "?" + query
		}

		args := []any{
			"request_id", RequestIDFrom(c),
			"user_id", UserIDFrom(c),
			"method", c.Request.Method,
			"path", fullPath,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"ip", c.ClientIP(),
		}

		if len(c.Errors) > 0 {
			args = append(args, "errors", c.Errors.String())
		}

		switch {
		case status >= 500:
			logger.Error("request completed", args...)
		case status >= 400:
			logger.Warn("request completed", args...)
		default:
			logger.Info("request completed", args...)
		}
	}
}
