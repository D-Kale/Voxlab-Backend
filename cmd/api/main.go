package main

import (
	"log"

	"github.com/voxlab/voxlab-backend/internal/config"
	"github.com/voxlab/voxlab-backend/internal/database"
	appHTTP "github.com/voxlab/voxlab-backend/internal/http"
	"github.com/voxlab/voxlab-backend/internal/storage"
)

// @title           Voxlab API
// @version         1.0
// @description     API for the Voxlab public speaking educational platform
// @termsOfService  http://voxlab.com/terms

// @contact.name   API Support
// @contact.email  support@voxlab.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:3000
// @BasePath  /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer " followed by your JWT token
func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	cfg := config.MustGetConfig()

	if err := database.ConnectPostgres(cfg); err != nil {
		log.Fatalf("Error connecting to PostgreSQL: %v", err)
	}

	if err := database.ConnectRedis(cfg); err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}

	if err := storage.InitStorage(cfg); err != nil {
		log.Fatalf("Error initializing storage: %v", err)
	}

	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	router := appHTTP.NewRouter(cfg)

	log.Printf("Server starting on port %s", cfg.AppPort)
	log.Printf("Swagger docs: http://localhost:%s/swagger/index.html", cfg.AppPort)

	if err := router.Run(cfg.AppPort); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
