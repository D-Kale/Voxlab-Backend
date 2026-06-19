package controllers

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/services"
)

const maxUploadSize = 2 << 20

// UploadResponse is returned by upload endpoints.
type UploadResponse struct {
	URL string `json:"url" example:"https://storage.example.com/uploads/track-1.webp"`
}

type UploadController struct {
	service *services.UploadService
}

func NewUploadController(service *services.UploadService) *UploadController {
	return &UploadController{service: service}
}

// @Summary      Upload track image
// @Description  Accepts an image file (JPEG/PNG), optimizes it to WebP (quality 80), and uploads to MinIO.
// @Description  The track's icon_url is updated automatically. The old image is deleted from storage.
// @Description  Max file size: 2MB. Image is resized to fit 1920x1080 while preserving aspect ratio.
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int    true  "Track ID"
// @Param        file formData  file   true  "Image file (JPEG/PNG, max 2MB)"
// @Success      200  {object}  resources.UploadFileResponse  "Imagen subida correctamente"
// @Failure      400  {object}  resources.BadRequestError     "ID de track inválido o archivo faltante"
// @Failure      401  {object}  resources.UnauthorizedError   "Token no proporcionado o inválido"
// @Failure      403  {object}  resources.ForbiddenError      "Solo administradores pueden subir imágenes"
// @Failure      500  {object}  resources.InternalServerError "Error al subir o procesar la imagen"
// @Router       /api/v1/upload/track/{id} [post]
func (h *UploadController) UploadTrackImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid track id"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large or unreadable"})
		return
	}

	result, err := h.service.UploadTrackImage(c.Request.Context(), id, bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// @Summary      Upload module image
// @Description  Same as track upload but for a module's image_url. WebP optimized, max 2MB, 1920x1080.
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int    true  "Module ID"
// @Param        file formData  file   true  "Image file (JPEG/PNG, max 2MB)"
// @Success      200  {object}  resources.UploadFileResponse  "Imagen subida correctamente"
// @Failure      400  {object}  resources.BadRequestError     "ID de módulo inválido o archivo faltante"
// @Failure      401  {object}  resources.UnauthorizedError   "Token no proporcionado o inválido"
// @Failure      403  {object}  resources.ForbiddenError      "Solo administradores pueden subir imágenes"
// @Failure      500  {object}  resources.InternalServerError "Error al subir o procesar la imagen"
// @Router       /api/v1/upload/module/{id} [post]
func (h *UploadController) UploadModuleImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module id"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large or unreadable"})
		return
	}

	result, err := h.service.UploadModuleImage(c.Request.Context(), id, bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// @Summary      Upload lesson image
// @Description  Same as track upload but for a lesson's image_url. WebP optimized, max 2MB, 1920x1080.
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int    true  "Lesson ID"
// @Param        file formData  file   true  "Image file (JPEG/PNG, max 2MB)"
// @Success      200  {object}  resources.UploadFileResponse  "Imagen subida correctamente"
// @Failure      400  {object}  resources.BadRequestError     "ID de lección inválido o archivo faltante"
// @Failure      401  {object}  resources.UnauthorizedError   "Token no proporcionado o inválido"
// @Failure      403  {object}  resources.ForbiddenError      "Solo administradores pueden subir imágenes"
// @Failure      500  {object}  resources.InternalServerError "Error al subir o procesar la imagen"
// @Router       /api/v1/upload/lesson/{id} [post]
func (h *UploadController) UploadLessonImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large or unreadable"})
		return
	}

	result, err := h.service.UploadLessonImage(c.Request.Context(), id, bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// @Summary      Upload user avatar
// @Description  Uploads and sets the authenticated user's avatar image. WebP optimized, max 2MB, 400x400.
// @Description  The old avatar is deleted from storage automatically.
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData  file   true  "Avatar image (JPEG/PNG, max 2MB)"
// @Success      200  {object}  resources.UploadFileResponse  "Avatar actualizado correctamente"
// @Failure      400  {object}  resources.BadRequestError     "Archivo faltante o excede el tamaño máximo"
// @Failure      401  {object}  resources.UnauthorizedError   "Token no proporcionado o inválido"
// @Failure      500  {object}  resources.InternalServerError "Error al subir o procesar el avatar"
// @Router       /api/v1/upload/avatar [post]
func (h *UploadController) UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large or unreadable"})
		return
	}

	result, err := h.service.UploadAvatar(c.Request.Context(), userID.(string), bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
