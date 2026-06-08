package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key" json:"id" swaggertype:"string"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	XP           int            `gorm:"type:int;default:0" json:"xp"`
	StreakDays   int            `gorm:"type:int;default:0" json:"streak_days"`
	Lives        int            `gorm:"type:int;default:5" json:"lives"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Titles       []UserTitle    `gorm:"foreignKey:UserID" json:"titles,omitempty"`
}

type ProgressStatus struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Name      string    `gorm:"type:varchar(50);unique;not null" json:"name"`
	ColorCode string    `gorm:"type:varchar(7);not null" json:"color_code"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type GamifiedTitle struct {
	ID                   int          `gorm:"primary_key" json:"id"`
	Name                 string       `gorm:"type:varchar(100);not null" json:"name"`
	Description          string       `gorm:"type:text" json:"description"`
	RequirementCondition string       `gorm:"type:varchar(255);not null" json:"requirement_condition"`
	CreatedAt            time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	UserTitles           []UserTitle  `gorm:"foreignKey:TitleID" json:"user_titles,omitempty"`
}

type UserTitle struct {
	UserID     uuid.UUID    `gorm:"type:uuid;primary_key" json:"user_id" swaggertype:"string"`
	TitleID    int          `gorm:"primary_key" json:"title_id"`
	IsEquipped bool         `gorm:"default:false" json:"is_equipped"`
	CreatedAt  time.Time    `gorm:"autoCreateTime" json:"created_at"`
	User       User         `gorm:"foreignKey:UserID" json:"-"`
	Title      GamifiedTitle `gorm:"foreignKey:TitleID" json:"title,omitempty"`
}
