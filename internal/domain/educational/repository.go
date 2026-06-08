package educational

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindAllTracks() ([]Track, error) {
	var tracks []Track
	err := r.db.Preload("Modules.Lessons.Lesson.Exercises").Find(&tracks).Error
	return tracks, err
}
