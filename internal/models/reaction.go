package models

import (
	"time"

	"github.com/google/uuid"
)

type ReactionType string

const (
	ReactionTypeLike      ReactionType = "like"
	ReactionTypeSupport   ReactionType = "support"
	ReactionTypeInspire   ReactionType = "inspire"
	ReactionTypeCelebrate ReactionType = "celebrate"
)

type UserReaction struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key" json:"id" swaggertype:"string"`
	SenderID     uuid.UUID    `gorm:"type:uuid;not null" json:"sender_id" swaggertype:"string"`
	ReceiverID   uuid.UUID    `gorm:"type:uuid;not null" json:"receiver_id" swaggertype:"string"`
	ReactionType ReactionType `gorm:"type:varchar(50);not null" json:"reaction_type"`
	CreatedAt    time.Time    `gorm:"autoCreateTime" json:"created_at"`
}
