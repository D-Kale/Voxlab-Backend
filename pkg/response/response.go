package response

import (
	"github.com/gin-gonic/gin"
)

// Response estructura genérica de respuesta
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success responde con éxito
func Success(c *gin.Context, data interface{}, message string) {
	c.JSON(200, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created responde con recurso creado
func Created(c *gin.Context, data interface{}, message string) {
	c.JSON(201, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error responde con error
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
	})
}

// BadRequest responde con error 400
func BadRequest(c *gin.Context, message string) {
	Error(c, 400, message)
}

// Unauthorized responde con error 401
func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message)
}

// NotFound responde con error 404
func NotFound(c *gin.Context, message string) {
	Error(c, 404, message)
}

// InternalError responde con error 500
func InternalError(c *gin.Context, message string) {
	Error(c, 500, message)
}