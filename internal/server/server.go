package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/voxlab/voxlab-backend/docs"

	"github.com/voxlab/voxlab-backend/internal/config"
	infraDB "github.com/voxlab/voxlab-backend/internal/infrastructure/database"
	"github.com/voxlab/voxlab-backend/internal/infrastructure/middleware"
	infraStorage "github.com/voxlab/voxlab-backend/internal/infrastructure/storage"

	"github.com/voxlab/voxlab-backend/internal/domain/auth"
	"github.com/voxlab/voxlab-backend/internal/domain/community"
	"github.com/voxlab/voxlab-backend/internal/domain/educational"
	"github.com/voxlab/voxlab-backend/internal/domain/user"
)

type Server struct {
	cfg      *config.Config
	router   *gin.Engine

	authH        *auth.Handler
	userH        *user.Handler
	educationalH *educational.Handler
	communityH   *community.Handler
}

func New() *Server {
	return &Server{}
}

func (s *Server) Init() error {
	if err := config.LoadConfig(); err != nil {
		return fmt.Errorf("error cargando configuración: %w", err)
	}
	s.cfg = config.MustGetConfig()

	if err := infraDB.ConnectPostgres(s.cfg); err != nil {
		return fmt.Errorf("error conectando a PostgreSQL: %w", err)
	}

	if err := infraDB.ConnectRedis(s.cfg); err != nil {
		return fmt.Errorf("error conectando a Redis: %w", err)
	}

	if err := infraStorage.InitMinio(s.cfg); err != nil {
		return fmt.Errorf("error inicializando MinIO: %w", err)
	}

	if err := infraDB.AutoMigrate(); err != nil {
		return fmt.Errorf("error migrando base de datos: %w", err)
	}

	s.initDependencies()
	s.initRouter()

	return nil
}

func (s *Server) initDependencies() {
	db := infraDB.GetDB()

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, db)
	s.userH = user.NewHandler(userSvc)

	authSvc := auth.NewService(userRepo, s.cfg.JWT.Secret)
	s.authH = auth.NewHandler(authSvc)

	eduRepo := educational.NewRepository(db)
	eduSvc := educational.NewService(eduRepo)
	s.educationalH = educational.NewHandler(eduSvc)

	comRepo := community.NewRepository(db)
	s.communityH = community.NewHandler(comRepo)
}

func (s *Server) initRouter() {
	gin.SetMode(gin.DebugMode)
	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.CORSMiddleware())

	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := s.router.Group("/api/v1")
	{
		api.GET("/health", s.HealthCheck)

		api.POST("/auth/login", s.authH.Login)

		api.GET("/tracks", s.educationalH.GetTracks)
	}
}

func (s *Server) Run() error {
	port := s.cfg.AppPort
	fmt.Printf("🚀 Servidor corriendo en puerto %s\n", port)
	fmt.Printf("📚 Documentación Swagger: http://localhost:%s/swagger/index.html\n", port)

	return s.router.Run(":" + port)
}

func (s *Server) HealthCheck(c *gin.Context) {
	// Note: This is a method on *Server to access any health check data if needed
	log.Println("Health check")
	c.JSON(200, gin.H{
		"status":    "ok",
		"timestamp": "TODO: add timestamp",
		"version":   "1.0.0",
	})
}
