package models

import "time"

type Track struct {
	ID          int      `gorm:"primary_key" json:"id"`
	Title       string   `gorm:"type:varchar(100);not null" json:"title"`
	Description string   `gorm:"type:text" json:"description"`
	IconURL     string   `gorm:"type:varchar(255)" json:"icon_url"`
	IsActive    bool     `gorm:"default:true" json:"is_active"`
	Modules     []Module `gorm:"foreignKey:TrackID" json:"modules,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Module struct {
	ID          int            `gorm:"primary_key" json:"id"`
	TrackID     int            `gorm:"not null" json:"track_id"`
	Title       string         `gorm:"type:varchar(100);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	ImageURL    string         `gorm:"type:varchar(512)" json:"image_url"`
	OrderIndex  int            `gorm:"not null" json:"order_index"`
	Lessons     []ModuleLesson `gorm:"foreignKey:ModuleID" json:"lessons,omitempty"`
	Track       Track          `gorm:"foreignKey:TrackID" json:"-"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

type Lesson struct {
	ID                   int              `gorm:"primary_key" json:"id"`
	Title                string           `gorm:"type:varchar(150);not null" json:"title"`
	Description          string           `gorm:"type:text" json:"description"`
	ImageURL             string           `gorm:"type:varchar(512)" json:"image_url"`
	EstimatedTimeSeconds int              `gorm:"not null" json:"estimated_time_seconds"`
	LessonExercises      []LessonExercise `gorm:"foreignKey:LessonID" json:"exercises,omitempty"`
	CreatedAt            time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

type ModuleLesson struct {
	ModuleID   int       `gorm:"primary_key" json:"module_id"`
	LessonID   int       `gorm:"primary_key" json:"lesson_id"`
	OrderIndex int       `gorm:"not null" json:"order_index"`
	Module     Module    `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Lesson     Lesson    `gorm:"foreignKey:LessonID" json:"lesson,omitempty"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type SharedLessonInfo struct {
	LessonID     int         `json:"lesson_id"`
	LessonTitle  string      `json:"lesson_title"`
	OtherModules []ModuleRef `json:"other_modules"`
}

type ModuleRef struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}
