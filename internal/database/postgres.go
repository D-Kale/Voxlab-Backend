package database

import (
	"fmt"
	"log"
	"os"

	"github.com/voxlab/voxlab-backend/internal/config"
	"github.com/voxlab/voxlab-backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectPostgres(cfg *config.Config) error {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.DB.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("connecting to PostgreSQL: %w", err)
	}

	log.Println("PostgreSQL connection established")
	return nil
}

func AutoMigrate() error {
	modelsList := []interface{}{
		&models.User{},
		&models.ProgressStatus{},
		&models.GamifiedTitle{},
		&models.UserTitle{},
		&models.Track{},
		&models.Module{},
		&models.Lesson{},
		&models.ModuleLesson{},
		&models.Exercise{},
		&models.UserReaction{},
		&models.UserProgress{},
	}

	for _, model := range modelsList {
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("migrating %T: %w", model, err)
		}
	}

	log.Println("Migrations completed")

	if err := runSeed(); err != nil {
		log.Printf("Seed warning: %v", err)
	}

	if err := SeedAdmin(config.MustGetConfig()); err != nil {
		log.Printf("Admin seed warning: %v", err)
	}

	return nil
}

func runSeed() error {
	seedPaths := []string{"database/seed.sql", "/app/database/seed.sql", "./database/seed.sql"}
	var seedData []byte
	var err error

	for _, path := range seedPaths {
		seedData, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if seedData == nil {
		return fmt.Errorf("seed.sql not found in any known path")
	}

	if err := DB.Exec(string(seedData)).Error; err != nil {
		return fmt.Errorf("executing seed: %w", err)
	}

	log.Println("Seed data inserted")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
