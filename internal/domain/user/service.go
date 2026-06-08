package user

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
	db   *gorm.DB
}

func NewService(repo *Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

func (s *Service) GetUserProfile(id string) (*User, error) {
	return s.repo.FindByID(id)
}

func (s *Service) AddXP(userID uuid.UUID, amount int) error {
	return s.db.Model(&User{}).Where("id = ?", userID).Update("xp", gorm.Expr("xp + ?", amount)).Error
}
