package repositories

import (
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

type TrackRepository struct {
	db *gorm.DB
}

func NewTrackRepository(db *gorm.DB) *TrackRepository {
	return &TrackRepository{db: db}
}

func (r *TrackRepository) FindAll() ([]models.Track, error) {
	var tracks []models.Track
	err := r.db.Preload("Modules.Lessons.Lesson.LessonExercises.Exercise").Find(&tracks).Error
	return tracks, err
}

func (r *TrackRepository) FindByID(id int) (*models.Track, error) {
	var track models.Track
	err := r.db.Preload("Modules.Lessons.Lesson.LessonExercises.Exercise").First(&track, id).Error
	if err != nil {
		return nil, err
	}
	return &track, nil
}

func (r *TrackRepository) Create(track *models.Track) error {
	return r.db.Create(track).Error
}

func (r *TrackRepository) Update(track *models.Track) error {
	return r.db.Save(track).Error
}

func (r *TrackRepository) Delete(id int) error {
	return r.db.Delete(&models.Track{}, id).Error
}
