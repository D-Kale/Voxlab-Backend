package repositories

import (
	"time"

	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

type ClientLogRepository struct {
	db *gorm.DB
}

func NewClientLogRepository(db *gorm.DB) *ClientLogRepository {
	return &ClientLogRepository{db: db}
}

func (r *ClientLogRepository) Create(log *models.ClientLog) error {
	return r.db.Create(log).Error
}

// ListRecent devuelve los logs más recientes, del más nuevo al más viejo.
// El admin los consulta para ver qué falló en producción sin Reproducir.
func (r *ClientLogRepository) ListRecent(level string, limit int) ([]models.ClientLog, error) {
	q := r.db.Order("created_at DESC").Limit(limit)
	if level != "" {
		q = q.Where("level = ?", level)
	}

	var logs []models.ClientLog
	if err := q.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// PurgeOlder borra logs con más de `age` de antigüedad para que la tabla
// no crezca sin límite. Se llama desde una tarea de mantenimiento.
func (r *ClientLogRepository) PurgeOlder(age time.Duration) (int64, error) {
	result := r.db.Where("created_at < ?", time.Now().Add(-age)).Delete(&models.ClientLog{})
	return result.RowsAffected, result.Error
}
