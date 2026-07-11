package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProgressRepository struct {
	db *gorm.DB
}

func NewProgressRepository(db *gorm.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

func (r *ProgressRepository) Upsert(progress *models.UserProgress) error {
	now := time.Now()
	progress.UpdatedAt = now

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "lesson_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "xp_earned", "completed_exercises", "completed_at", "updated_at"}),
	}).Create(progress).Error
}

func (r *ProgressRepository) Update(progress *models.UserProgress) error {
	progress.UpdatedAt = time.Now()
	return r.db.Save(progress).Error
}

func (r *ProgressRepository) FindByUserAndLesson(userID uuid.UUID, lessonID int) (*models.UserProgress, error) {
	var progress models.UserProgress
	err := r.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *ProgressRepository) FindByUserAndLessons(userID uuid.UUID, lessonIDs []int) ([]models.UserProgress, error) {
	var progress []models.UserProgress
	err := r.db.Where("user_id = ? AND lesson_id IN ?", userID, lessonIDs).Find(&progress).Error
	return progress, err
}

func (r *ProgressRepository) FindAllByUser(userID uuid.UUID) ([]models.UserProgress, error) {
	var progress []models.UserProgress
	err := r.db.Where("user_id = ?", userID).Order("lesson_id asc").Find(&progress).Error
	return progress, err
}

func (r *ProgressRepository) DeleteByUserAndLesson(userID uuid.UUID, lessonID int) error {
	return r.db.Where("user_id = ? AND lesson_id = ?", userID, lessonID).Delete(&models.UserProgress{}).Error
}

func (r *ProgressRepository) FindLatestByUser(userID uuid.UUID) (*models.UserProgress, error) {
	var progress models.UserProgress
	err := r.db.Where("user_id = ?", userID).Order("updated_at desc").First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}
