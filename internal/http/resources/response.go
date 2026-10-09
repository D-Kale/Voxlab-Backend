// Package resources contiene los helpers de respuesta HTTP compartidos.
package resources

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/http/middleware"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	// RequestID se devuelve en los errores para que el frontend pueda
	// correlacionar la falla del cliente con la línea del server en los logs.
	RequestID string `json:"request_id,omitempty"`
}

func Success(c *gin.Context, data any, message string) {
	c.JSON(200, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, data any, message string) {
	c.JSON(201, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success:   false,
		Error:     message,
		RequestID: middleware.RequestIDFrom(c),
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, 400, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, 404, message)
}

// InternalError responde 500 y REGISTRA la causa real en los logs.
//
// Antes los controllers escribían directamente
// c.JSON(500, gin.H{"error": "Failed to ..."}) y descartaban el err, así que
// la causa nunca quedaba registrada. Ahora el servidor la loguea (con
// request_id) y el cliente sigue recibiendo un mensaje genérico y seguro.
func InternalError(c *gin.Context, err error, clientMessage string) {
	middleware.LogRequestError(c, err, clientMessage)
	Error(c, http.StatusInternalServerError, clientMessage)
}

// Fail responde con el status resuelto y registra el error SOLO si es 5xx.
//
// Existe para los handlers que arman el status a mano: arrancan en 500 y lo
// bajan según el string del error (404 / 409 / 403). Cuando el status final es
// un error del cliente no hay nada que loguear del lado del servidor; cuando
// cae en 500 sí, y antes esa causa se perdía.
func Fail(c *gin.Context, err error, statusCode int, message string) {
	if statusCode >= 500 {
		middleware.LogRequestError(c, err, message)
	}
	Error(c, statusCode, message)
}
