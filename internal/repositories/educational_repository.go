package repositories

import (
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

type EducationalRepository struct {
	db *gorm.DB
}

func NewEducationalRepository(db *gorm.DB) *EducationalRepository {
	return &EducationalRepository{db: db}
}

func (r *EducationalRepository) FindAllTracks() ([]models.Track, error) {
	var tracks []models.Track
	err := r.db.Preload("Modules.Lessons.Lesson.Exercises").Find(&tracks).Error
	return tracks, err
}
