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

// withLearningPath preloads modules, lessons and exercises sorted by their
// order_index, so the app's map follows the order set with the reorder
// endpoints (a bare Preload returns rows in arbitrary order).
func withLearningPath(db *gorm.DB) *gorm.DB {
	byOrder := func(db *gorm.DB) *gorm.DB { return db.Order("order_index asc") }
	return db.
		Preload("Modules", byOrder).
		Preload("Modules.Lessons", byOrder).
		Preload("Modules.Lessons.Lesson").
		Preload("Modules.Lessons.Lesson.LessonExercises", byOrder).
		Preload("Modules.Lessons.Lesson.LessonExercises.Exercise")
}

func (r *TrackRepository) FindAll() ([]models.Track, error) {
	var tracks []models.Track
	err := withLearningPath(r.db).Order("order_index asc, id asc").Find(&tracks).Error
	return tracks, err
}

func (r *TrackRepository) FindByID(id int) (*models.Track, error) {
	var track models.Track
	err := withLearningPath(r.db).First(&track, id).Error
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

type TrackOrderItem struct {
	ID         int `json:"id"`
	OrderIndex int `json:"order_index"`
}

func (r *TrackRepository) BatchUpdateOrder(items []TrackOrderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Model(&models.Track{}).Where("id = ?", item.ID).Update("order_index", item.OrderIndex).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
