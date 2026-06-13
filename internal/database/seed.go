package database

import (
	"log"

	"github.com/google/uuid"
	"github.com/voxlab/voxlab-backend/internal/config"
	"github.com/voxlab/voxlab-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedAdmin(cfg *config.Config) error {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		log.Println("Admin user already exists, skipping seed")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &models.User{
		ID:           uuid.New(),
		Name:         "Admin",
		Email:        cfg.Admin.Email,
		PasswordHash: string(hash),
		Role:         "admin",
	}

	if err := DB.Create(admin).Error; err != nil {
		return err
	}

	log.Printf("Admin user created: %s", cfg.Admin.Email)
	return nil
}
