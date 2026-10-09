package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/logger"
)

const (
	// RequestIDKey es la clave bajo la que el request_id vive en el context de gin.
	RequestIDKey = "request_id"
	// RequestIDHeader es el header que el frontend lee para correlacionar.
	RequestIDHeader = "X-Request-ID"
)

// RequestID asigna un ID único a cada request.
//
// Respeta el header entrante para que un retry del mismo cliente conserve la
// correlación. El ID se devuelve en el header de respuesta, lo que permite
// pegarlo en un bug report y encontrar la línea exacta en docker logs.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		c.Set(RequestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)

		c.Next()
	}
}

// RequestIDFrom devuelve el request_id del context, o "-" si no hay.
func RequestIDFrom(c *gin.Context) string {
	if id, ok := c.Get(RequestIDKey); ok {
		if s, ok := id.(string); ok && s != "" {
			return s
		}
	}
	return "-"
}

// UserIDFrom devuelve el user_id que AuthMiddleware inyectó, o "-" si la
// request es anónima. Nunca loguea passwords ni el token completo.
func UserIDFrom(c *gin.Context) string {
	if id, ok := c.Get("user_id"); ok {
		if s, ok := id.(string); ok && s != "" {
			return s
		}
	}
	return "-"
}

// LogRequestError loguea un error que va a devolverse al cliente.
//
// El punto central de este middleware: los controllers llamaban
// c.JSON(500, gin.H{"error": "Failed to ..."}) descartando el err original,
// así que la causa real no quedaba registrada en ningún lado. Ahora la causa
// queda en los logs del servidor y el cliente sigue viendo un mensaje seguro.
func LogRequestError(c *gin.Context, err error, clientMessage string) {
	args := []any{
		"request_id", RequestIDFrom(c),
		"user_id", UserIDFrom(c),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"client_message", clientMessage,
	}
	if err != nil {
		args = append(args, "error", err.Error())
	}

	logger.Error("request failed", args...)
}

// LogRequestWarn loguea una condición anómala que no corta la request.
func LogRequestWarn(c *gin.Context, message string, args ...any) {
	base := []any{
		"request_id", RequestIDFrom(c),
		"user_id", UserIDFrom(c),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	}
	logger.Warn(message, append(base, args...)...)
}
