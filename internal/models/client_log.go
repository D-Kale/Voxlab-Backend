package models

import (
	"time"

	"encoding/json"

	"github.com/google/uuid"
)

// ClientLog captura un error o warning del navegador.
//
// El frontend no tenía ningún sink: los errores vivían en la consola del
// developer y se perdían al recargar. Con este endpoint + src/lib/logger.ts
// el equipo puede consultar qué falló en producción.
//
// La fila se persiste Y se escribe en el log del servidor, así también queda
// en docker logs / cualquier agregador.
type ClientLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id" swaggertype:"string"`
	Level     string    `gorm:"type:varchar(20);not null;index" json:"level"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Source    string    `gorm:"type:varchar(255)" json:"source"`
	URL       string    `gorm:"type:varchar(512)" json:"url"`
	UserAgent string    `gorm:"type:varchar(512)" json:"user_agent"`
	Stack     string    `gorm:"type:text" json:"stack,omitempty"`
	// RequestID correlaciona con la línea del access log del servidor que
	// originó el error del cliente.
	RequestID string `gorm:"type:varchar(64);index" json:"request_id,omitempty"`
	UserID    string `gorm:"type:varchar(64);index" json:"user_id,omitempty"`
	// Context lleva datos extra (status HTTP, método, payload del request).
	// json.RawMessage y no datatypes.JSON porque es lo que ya usan Exercise
	// y Progress, y evita sumar una dependencia.
	Context   json.RawMessage `gorm:"type:jsonb" json:"context,omitempty"`
	CreatedAt time.Time       `gorm:"autoCreateTime;index" json:"created_at"`
}

// TableName evita que GORM pluralice a "client_logs" si algún día cambia el
// criterio; se fija explícitamente para que las queries manuales no cambien.
func (ClientLog) TableName() string {
	return "client_logs"
}
