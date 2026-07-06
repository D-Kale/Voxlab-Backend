package resources

import (
	"github.com/voxlab/voxlab-backend/internal/analyzer"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/services"
)

// ============================================================================
// Base Responses — embedded in concrete response DTOs
// ============================================================================

// BaseResponse is the standard wrapper for all API responses.
// It provides the "success" field and an optional "message".
type BaseResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message,omitempty" example:"Operación realizada con éxito"`
}

// ============================================================================
// Error Responses
// ============================================================================

// BadRequestError — 400
type BadRequestError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Solicitud inválida — verifique los campos enviados"`
}

// UnauthorizedError — 401
type UnauthorizedError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Token de autenticación no proporcionado o inválido"`
}

// ForbiddenError — 403
type ForbiddenError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"No tiene permisos para realizar esta acción"`
}

// NotFoundError — 404
type NotFoundError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Recurso no encontrado"`
}

// ConflictError — 409
type ConflictError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"El recurso ya existe — conflicto con datos existentes"`
}

// InternalServerError — 500
type InternalServerError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Error interno del servidor"`
}

// NotImplementedError — 501
type NotImplementedError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Funcionalidad no implementada aún"`
}

// ServiceUnavailableError — 502/503
type ServiceUnavailableError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"Servicio temporalmente no disponible — intente más tarde"`
}

// ============================================================================
// AUTH — /api/v1/auth/*
// ============================================================================

// LoginResponseData wraps the login response.
// NOTE: the actual endpoint returns services.LoginResponse directly (no success wrapper).
type LoginResponseData services.LoginResponse

// RegisterResponseData wraps the register response.
// Same as LoginResponseData — returns services.LoginResponse directly.
type RegisterResponseData services.LoginResponse

// LogoutResponse represents a successful logout.
type LogoutResponse struct {
	BaseResponse
}

// UserProfileResponse represents a user profile (me / profile endpoints).
type UserProfileResponse struct {
	BaseResponse
	Data services.UserData `json:"data"`
}

// ============================================================================
// TRACKS — /api/v1/tracks/*
// ============================================================================

type ListTracksResponse struct {
	BaseResponse
	Data []models.Track `json:"data"`
}

type GetTrackResponse struct {
	BaseResponse
	Data models.Track `json:"data"`
}

type CreateTrackResponse struct {
	BaseResponse
	Data models.Track `json:"data"`
}

type UpdateTrackResponse struct {
	BaseResponse
	Data models.Track `json:"data"`
}

type DeleteTrackResponse struct {
	BaseResponse
}

// ============================================================================
// MODULES — /api/v1/modules/* and /api/v1/tracks/:id/modules
// ============================================================================

type ListModulesResponse struct {
	BaseResponse
	Data []models.Module `json:"data"`
}

type GetModuleResponse struct {
	BaseResponse
	Data models.Module `json:"data"`
}

type CreateModuleResponse struct {
	BaseResponse
	Data models.Module `json:"data"`
}

type UpdateModuleResponse struct {
	BaseResponse
	Data models.Module `json:"data"`
}

type DeleteModuleResponse struct {
	BaseResponse
}

type LinkLessonResponse struct {
	BaseResponse
}

type UnlinkLessonResponse struct {
	BaseResponse
}

type ReorderLessonsResponse struct {
	BaseResponse
}

// ============================================================================
// LESSON ⇄ EXERCISE links — /api/v1/lessons/:id/exercises and /api/v1/exercises/:id/lessons
// ============================================================================

type LinkExerciseResponse struct {
	BaseResponse
}

type UnlinkExerciseResponse struct {
	BaseResponse
}

type ReorderExerciseResponse struct {
	BaseResponse
}

type ReorderExercisesResponse struct {
	BaseResponse
}

type GetModulesByLessonResponse struct {
	BaseResponse
	Data []models.ModuleLesson `json:"data"`
}

type GetLessonsByExerciseResponse struct {
	BaseResponse
	Data []models.LessonExercise `json:"data"`
}

type GetExercisesByLessonResponse struct {
	BaseResponse
	Data []models.LessonExercise `json:"data"`
}

// ============================================================================
// LESSONS — /api/v1/lessons/* and /api/v1/modules/:id/lessons
// ============================================================================

type ListLessonsResponse struct {
	BaseResponse
	Data []models.Lesson `json:"data"`
}

type GetLessonResponse struct {
	BaseResponse
	Data models.Lesson `json:"data"`
}

type CreateLessonResponse struct {
	BaseResponse
	Data models.Lesson `json:"data"`
}

type UpdateLessonResponse struct {
	BaseResponse
	Data models.Lesson `json:"data"`
}

type DeleteLessonResponse struct {
	BaseResponse
}

// ============================================================================
// EXERCISES — /api/v1/exercises/* and /api/v1/lessons/:id/exercises
// ============================================================================

type ListExercisesResponse struct {
	BaseResponse
	Data []models.Exercise `json:"data"`
}

type GetExerciseResponse struct {
	BaseResponse
	Data models.Exercise `json:"data"`
}

type CreateExerciseResponse struct {
	BaseResponse
	Data models.Exercise `json:"data"`
}

type UpdateExerciseResponse struct {
	BaseResponse
	Data models.Exercise `json:"data"`
}

type DeleteExerciseResponse struct {
	BaseResponse
}

type RequirementCatalogResponse struct {
	BaseResponse
	Data []models.RequirementCatalogItem `json:"data"`
}

type AnalyzeTextResponse struct {
	BaseResponse
	Data analyzer.AnalyzeResponse `json:"data"`
}

// ============================================================================
// PROGRESS — /api/v1/progress/*
// ============================================================================

type ListProgressResponse struct {
	BaseResponse
	Data []models.UserProgress `json:"data"`
}

type CompleteProgressResponse struct {
	BaseResponse
	Data models.UserProgress `json:"data"`
}

type UpdateProgressResponse struct {
	BaseResponse
	Data models.UserProgress `json:"data"`
}

// ============================================================================
// USERS — /api/v1/users/*
// ============================================================================

type ListUsersResponse struct {
	BaseResponse
	Data []services.AdminUserData `json:"data"`
}

type GetUserResponse struct {
	BaseResponse
	Data services.AdminUserData `json:"data"`
}

type UpdateUserResponse struct {
	BaseResponse
	Data services.AdminUserData `json:"data"`
}

type DeleteUserResponse struct {
	BaseResponse
}

// ============================================================================
// UPLOAD — /api/v1/upload/*
// ============================================================================

type UploadFileResponse struct {
	BaseResponse
	Data struct {
		URL string `json:"url" example:"https://storage.voxlab.com/uploads/track-1.webp"`
	} `json:"data"`
}

// ============================================================================
// REACTIONS — /api/v1/reactions (placeholder / not yet implemented)
// ============================================================================

type NotImplementedResponse struct {
	Message string `json:"message" example:"not implemented"`
}
