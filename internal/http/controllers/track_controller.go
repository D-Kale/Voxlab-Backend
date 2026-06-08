package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type TrackController struct {
	service *services.TrackService
}

func NewTrackController(service *services.TrackService) *TrackController {
	return &TrackController{service: service}
}

// GetTracks      godoc
// @Summary      List all educational tracks
// @Description  Returns ALL available tracks with their modules, lessons, and exercises nested inside.
// @Description  This is the main endpoint for the course catalog — frontends should call this once
// @Description  to build the full navigation tree. Each track contains modules, each module contains
// @Description  lessons (through the pivot table), and each lesson contains exercises.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Tracks (Educational Content)
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Success: { success: true, data: Track[] }"
// @Failure      500  {object}  map[string]interface{}  "Server error"
// @Router       /api/v1/tracks [get]
func (h *TrackController) GetTracks(c *gin.Context) {
	tracks, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tracks"})
		return
	}

	if tracks == nil {
		tracks = []models.Track{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tracks,
	})
}

// GetTrack       godoc
// @Summary      Get a single track by ID
// @Description  Returns one track with its nested modules, lessons, and exercises.
// @Description  Use this when you need to reload or fetch details for a specific track.
// @Description
// @Description  🔓 Public — no authentication required.
// @Tags         Tracks (Educational Content)
// @Produce      json
// @Param        id   path      int  true  "Track ID (e.g. 1)"
// @Success      200  {object}  map[string]interface{}  "Success: { success: true, data: Track }"
// @Failure      404  {object}  map[string]interface{}  "Track not found"
// @Router       /api/v1/tracks/{id} [get]
func (h *TrackController) GetTrack(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	track, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Track not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    track,
	})
}

// CreateTrack    godoc
// @Summary      Create a new track
// @Description  Adds a new educational track (course) to the catalog.
// @Description  After creating a track, you can add modules to it using POST /api/v1/modules.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Tracks (Educational Content)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      object{title=string,description=string,icon_url=string}  true  "Track data"
// @Success      201  {object}  map[string]interface{}  "Created: { success: true, data: Track }"
// @Failure      400  {object}  map[string]interface{}  "Validation error"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Router       /api/v1/tracks [post]
func (h *TrackController) CreateTrack(c *gin.Context) {
	var track models.Track
	if err := c.ShouldBindJSON(&track); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if track.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	if err := h.service.Create(&track); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create track"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    track,
	})
}

// UpdateTrack    godoc
// @Summary      Update an existing track
// @Description  Modifies the title, description, or icon of a track.
// @Description  Send only the fields you want to update.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Tracks (Educational Content)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path    int                                     true  "Track ID (e.g. 1)"
// @Param        request  body    object{title=string,description=string,icon_url=string}  true  "Track fields to update"
// @Success      200  {object}  map[string]interface{}  "Updated: { success: true, data: Track }"
// @Failure      400  {object}  map[string]interface{}  "Validation error"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      404  {object}  map[string]interface{}  "Track not found"
// @Router       /api/v1/tracks/{id} [put]
func (h *TrackController) UpdateTrack(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	existing, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Track not found"})
		return
	}

	var input models.Track
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if input.Title != "" {
		existing.Title = input.Title
	}
	if input.Description != "" {
		existing.Description = input.Description
	}
	if input.IconURL != "" {
		existing.IconURL = input.IconURL
	}

	if err := h.service.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update track"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    existing,
	})
}

// DeleteTrack    godoc
// @Summary      Delete a track
// @Description  Permanently removes a track and all its associated modules and module-lesson links.
// @Description  ⚠️ This action cannot be undone.
// @Description
// @Description  🔒 Requires JWT token (Authorization: Bearer <token>)
// @Tags         Tracks (Educational Content)
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Track ID (e.g. 1)"
// @Success      200  {object}  map[string]interface{}  "Deleted: { success: true, message: string }"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      404  {object}  map[string]interface{}  "Track not found"
// @Router       /api/v1/tracks/{id} [delete]
func (h *TrackController) DeleteTrack(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	if _, err := h.service.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Track not found"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete track"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Track deleted successfully",
	})
}
