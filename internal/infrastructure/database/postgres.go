package database

import (
	"fmt"
	"log"
	"os"

	"github.com/voxlab/voxlab-backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/voxlab/voxlab-backend/internal/domain/community"
	"github.com/voxlab/voxlab-backend/internal/domain/educational"
	"github.com/voxlab/voxlab-backend/internal/domain/user"
)

var DB *gorm.DB

func ConnectPostgres(cfg *config.Config) error {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.DB.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("error conectando a PostgreSQL: %w", err)
	}

	log.Println("✅ Conexión a PostgreSQL establecida")
	return nil
}

func AutoMigrate() error {
	models := []interface{}{
		&user.User{},
		&user.ProgressStatus{},
		&user.GamifiedTitle{},
		&user.UserTitle{},
		&educational.Track{},
		&educational.Module{},
		&educational.Lesson{},
		&educational.ModuleLesson{},
		&educational.Exercise{},
		&community.UserReaction{},
	}

	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("error migrando %T: %w", model, err)
		}
	}

	log.Println("✅ Migraciones completadas exitosamente")

	if err := runSeed(); err != nil {
		log.Printf("⚠️  Advertencia ejecutando seed: %v", err)
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
		return fmt.Errorf("no se encontró el archivo seed.sql en ninguna ruta conocida")
	}

	if err := DB.Exec(string(seedData)).Error; err != nil {
		return fmt.Errorf("error ejecutando seed: %w", err)
	}

	log.Println("✅ Seed data insertado exitosamente")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
