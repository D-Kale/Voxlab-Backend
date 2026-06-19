package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

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
		&models.LessonExercise{},
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

	if err := runMigrations(); err != nil {
		return fmt.Errorf("running SQL migrations: %w", err)
	}

	if err := runSeed(); err != nil {
		log.Printf("Seed warning: %v", err)
	}

	if err := SeedAdmin(config.MustGetConfig()); err != nil {
		log.Printf("Admin seed warning: %v", err)
	}

	return nil
}

func runMigrations() error {
	searchPaths := []string{
		"database/migrations",
		"/app/database/migrations",
		"./database/migrations",
	}

	var migrationDir string
	for _, path := range searchPaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			migrationDir = path
			break
		}
	}

	if migrationDir == "" {
		log.Println("No migrations directory found, skipping SQL migrations")
		return nil
	}

	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	var sqlFiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			sqlFiles = append(sqlFiles, filepath.Join(migrationDir, e.Name()))
		}
	}
	sort.Strings(sqlFiles)

	for _, path := range sqlFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", path, err)
		}

		log.Printf("Running migration: %s", filepath.Base(path))
		if err := DB.Exec(string(data)).Error; err != nil {
			return fmt.Errorf("executing migration %s: %w", filepath.Base(path), err)
		}
	}

	log.Printf("SQL migrations completed (%d files)", len(sqlFiles))
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
