package educational

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Track struct {
	ID          int       `gorm:"primary_key" json:"id"`
	Title       string    `gorm:"type:varchar(100);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	IconURL     string    `gorm:"type:varchar(255)" json:"icon_url"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	Modules     []Module  `gorm:"foreignKey:TrackID" json:"modules,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Module struct {
	ID          int            `gorm:"primary_key" json:"id"`
	TrackID     int            `gorm:"not null" json:"track_id"`
	Title       string         `gorm:"type:varchar(100);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	OrderIndex  int            `gorm:"not null" json:"order_index"`
	Lessons     []ModuleLesson `gorm:"foreignKey:ModuleID" json:"lessons,omitempty"`
	Track       Track          `gorm:"foreignKey:TrackID" json:"-"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

type Lesson struct {
	ID                   int        `gorm:"primary_key" json:"id"`
	Title                string     `gorm:"type:varchar(150);not null" json:"title"`
	Description          string     `gorm:"type:text" json:"description"`
	EstimatedTimeSeconds int        `gorm:"not null" json:"estimated_time_seconds"`
	Exercises            []Exercise `gorm:"foreignKey:LessonID" json:"exercises,omitempty"`
	CreatedAt            time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

type ModuleLesson struct {
	ModuleID   int       `gorm:"primary_key" json:"module_id"`
	LessonID   int       `gorm:"primary_key" json:"lesson_id"`
	OrderIndex int       `gorm:"not null" json:"order_index"`
	Module     Module    `gorm:"foreignKey:ModuleID" json:"-"`
	Lesson     Lesson    `gorm:"foreignKey:LessonID" json:"lesson,omitempty"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type ExerciseType string

const (
	ExerciseTypeReading        ExerciseType = "reading"
	ExerciseTypeQuiz           ExerciseType = "quiz"
	ExerciseTypeAudio          ExerciseType = "audio"
	ExerciseTypeOratoryMinigame ExerciseType = "oratory_minigame"
	ExerciseTypeVideo          ExerciseType = "video"
)

type Exercise struct {
	ID         uuid.UUID       `gorm:"type:uuid;primary_key" json:"id" swaggertype:"string"`
	LessonID   int             `gorm:"not null" json:"lesson_id"`
	Type       ExerciseType    `gorm:"type:varchar(50);not null" json:"type"`
	OrderIndex int             `gorm:"not null" json:"order_index"`
	Content    json.RawMessage `gorm:"type:jsonb;not null" json:"content"`
	Lesson     Lesson          `gorm:"foreignKey:LessonID" json:"-"`
	CreatedAt  time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

type ExerciseContentQuiz struct {
	Question     string   `json:"question"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
	Explanation  string   `json:"explanation"`
	Points       int      `json:"points"`
}

type ExerciseContentOratoryMinigame struct {
	Prompt           string   `json:"prompt"`
	Topic            string   `json:"topic"`
	DurationSeconds  int      `json:"duration_seconds"`
	MinDurationSeconds int    `json:"min_duration_seconds"`
	Requirements     []string `json:"requirements"`
	Points           int      `json:"points"`
}

type ExerciseContentReading struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	ReadingTime int    `json:"reading_time_seconds"`
	Points      int    `json:"points"`
}
