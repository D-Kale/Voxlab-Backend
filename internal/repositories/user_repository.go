package repositories

import (
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var u models.User
	err := r.db.First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Order("created_at desc").Find(&users).Error
	return users, err
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.User{}).Error
}

// userDataTables hold rows that belong to a user. Some have no foreign key
// to users, so they are cleared explicitly before the user row goes.
var userDataTables = []string{
	"client_logs",
	"exercise_attempts",
	"user_exercise_progress",
	"user_exercise_progresses",
	"user_progress",
	"user_titles",
}

// HardDelete permanently removes a user and everything stored about them
// (account deletion requested by the user; Google Play requires it).
// Unlike Delete it does not soft-delete: the email and name are gone.
func (r *UserRepository) HardDelete(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, table := range userDataTables {
			if !tx.Migrator().HasTable(table) {
				continue
			}
			if err := tx.Exec("DELETE FROM "+table+" WHERE user_id = ?", id).Error; err != nil {
				return err
			}
		}
		return tx.Unscoped().Where("id = ?", id).Delete(&models.User{}).Error
	})
}

func (r *UserRepository) AddXP(id string, xp int) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).
		UpdateColumn("xp", gorm.Expr("xp + ?", xp)).Error
}

func (r *UserRepository) FindLeaderboard(limit int) ([]models.User, error) {
	var users []models.User
	err := r.db.Order("xp desc").Limit(limit).Find(&users).Error
	return users, err
}

func (r *UserRepository) CountByXPGreaterThan(xp int) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("xp > ?", xp).Count(&count).Error
	return count, err
}
