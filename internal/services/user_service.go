package services

import (
	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/models"
	"github.com/voxlab/voxlab-backend/internal/repositories"
	"gorm.io/gorm"
)

type UserService struct {
	repo *repositories.UserRepository
	db   *gorm.DB
}

func NewUserService(repo *repositories.UserRepository, db *gorm.DB) *UserService {
	return &UserService{repo: repo, db: db}
}

func (s *UserService) GetUserProfile(id string) (*models.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) AddXP(userID uuid.UUID, amount int) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("xp", gorm.Expr("xp + ?", amount)).Error
}
