package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/voxlab/voxlab-backend/internal/http/middleware"
	"github.com/voxlab/voxlab-backend/internal/http/resources"
	"github.com/voxlab/voxlab-backend/internal/logger"
	"github.com/voxlab/voxlab-backend/internal/repositories"
	"github.com/voxlab/voxlab-backend/internal/services"
)

const maxClientLogsPerRequest = 50

type ClientLogController struct {
	service *services.ClientLogService
	repo    *repositories.ClientLogRepository
}

func NewClientLogController(service *services.ClientLogService, repo *repositories.ClientLogRepository) *ClientLogController {
	return &ClientLogController{service: service, repo: repo}
}

// IngestBatchRequest es el payload que envía src/lib/logger.ts.
type IngestBatchRequest struct {
	Logs []services.ClientLogInput `json:"logs"`
}

// Ingest     godoc
// @Summary      Ingerir logs del cliente
// @Description  Recibe un lote de errores/warnings del navegador, los persiste y los escribe en el log del servidor.
// @Description  El `request_id` de cada entrada permite correlacionar con la línea del access log que la originó.
// @Tags         Logs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      IngestBatchRequest  true  "Lote de logs del cliente"
// @Success      202  {object}  resources.SuccessResponse  "Lote aceptado"
// @Failure      400  {object}  resources.BadRequestError  "Payload inválido"
// @Failure      401  {object}  resources.UnauthorizedError  "No autenticado"
// @Router       /api/v1/logs [post]
func (h *ClientLogController) Ingest(c *gin.Context) {
	var req IngestBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.LogRequestWarn(c, "invalid client log batch", "error", err.Error())
		resources.BadRequest(c, "Invalid request body")
		return
	}

	if len(req.Logs) == 0 {
		resources.Success(c, gin.H{"ingested": 0}, "Nothing to ingest")
		return
	}

	if len(req.Logs) > maxClientLogsPerRequest {
		// received se calcula antes de truncar: si se leyera después daría
		// siempre el mismo número y el warning no serviría para nada.
		received := len(req.Logs)
		req.Logs = req.Logs[:maxClientLogsPerRequest]
		logger.Warn("client log batch truncated",
			"request_id", middleware.RequestIDFrom(c),
			"received", received,
			"ingested", len(req.Logs),
			"max", maxClientLogsPerRequest,
		)
	}

	userID := middleware.UserIDFrom(c)
	userAgent := c.GetHeader("User-Agent")
	requestID := middleware.RequestIDFrom(c)

	for _, entry := range req.Logs {
		// El request_id del batch es la correlación más específica; si el
		// frontend no lo mandó, usamos el de esta request.
		if entry.RequestID == "" {
			entry.RequestID = requestID
		}
		if entry.UserAgent == "" {
			entry.UserAgent = userAgent
		}
		h.service.Record(entry, userID)
	}

	resources.Success(c, gin.H{"ingested": len(req.Logs)}, "Logs ingested")
}

// ListRecent    godoc
// @Summary      Listar logs recientes del cliente
// @Description  Devuelve los últimos errores/warnings del navegador. Solo admin.
// @Tags         Logs
// @Produce      json
// @Security     BearerAuth
// @Param        level  query     string  false  "Filtrar por nivel: error | warn | info"
// @Param        limit  query     int     false  "Máximo de resultados (default 100, max 500)"
// @Success      200  {object}  resources.SuccessResponse  "Lista de logs"
// @Failure      401  {object}  resources.UnauthorizedError  "No autenticado"
// @Failure      403  {object}  resources.ForbiddenError  "Se requiere rol admin"
// @Failure      500  {object}  resources.InternalServerError  "Error al obtener los logs"
// @Router       /api/v1/logs [get]
func (h *ClientLogController) ListRecent(c *gin.Context) {
	limit := 100
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	level := c.Query("level")

	logs, err := h.repo.ListRecent(level, limit)
	if err != nil {
		resources.InternalError(c, err, "Failed to fetch client logs")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"meta": gin.H{
			"count": len(logs),
			"level": level,
			"limit": limit,
		},
	})
}
