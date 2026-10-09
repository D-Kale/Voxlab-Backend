package main

import (
	"log"
	"os"

	"github.com/voxlab/voxlab-backend/internal/config"
	"github.com/voxlab/voxlab-backend/internal/database"
	appHTTP "github.com/voxlab/voxlab-backend/internal/http"
	"github.com/voxlab/voxlab-backend/internal/logger"
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

// @tag.name         Auth
// @tag.description  Authentication & user profile endpoints
// @tag.name         Tracks
// @tag.description  Track (course) management endpoints
// @tag.name         Modules
// @tag.description  Module management endpoints within tracks
// @tag.name         Lessons
// @tag.description  Lesson management endpoints within modules
// @tag.name         Exercises
// @tag.description  Exercise management within lessons + text analysis
// @tag.name         Progress
// @tag.description  User progress tracking (lesson completion, XP)
// @tag.name         Upload
// @tag.description  File upload endpoints (images, avatars)
// @tag.name         Users
// @tag.description  Admin user management endpoints
// @tag.name         Reactions
// @tag.description  Feedback & reactions on exercises
// @tag.name         System
// @tag.description  Health check and system endpoints
// @tag.name         Docs
// @tag.description  Swagger documentation endpoints
func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	cfg := config.MustGetConfig()

	// El logger se inicializa acá, antes de cualquier otra cosa, para que los
	// fallos de Postgres/Redis/MinIO ya queden en formato estructurado.
	logger.Init(cfg)

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

	logger.Info("server starting",
		"port", cfg.AppPort,
		"env", cfg.AppEnv,
		"swagger", "http://localhost:"+cfg.AppPort+"/swagger/index.html",
	)

	if err := router.Run(cfg.AppPort); err != nil {
		logger.Error("server error", "error", err.Error())
		os.Exit(1)
	}
}
