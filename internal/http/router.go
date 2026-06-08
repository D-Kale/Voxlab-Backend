package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/voxlab/voxlab-backend/docs"

	"github.com/voxlab/voxlab-backend/internal/config"
	"github.com/voxlab/voxlab-backend/internal/database"
	"github.com/voxlab/voxlab-backend/internal/http/controllers"
	"github.com/voxlab/voxlab-backend/internal/http/middleware"
	"github.com/voxlab/voxlab-backend/internal/repositories"
	"github.com/voxlab/voxlab-backend/internal/services"
)

type Router struct {
	engine *gin.Engine
	cfg    *config.Config

	health   *controllers.HealthController
	auth     *controllers.AuthController
	track    *controllers.TrackController
	module   *controllers.ModuleController
	lesson   *controllers.LessonController
	exercise *controllers.ExerciseController
	progress *controllers.ProgressController
	reaction *controllers.ReactionController
}

func NewRouter(cfg *config.Config) *Router {
	r := &Router{cfg: cfg}
	r.initDependencies()
	r.initEngine()
	return r
}

func (r *Router) initDependencies() {
	db := database.GetDB()

	userRepo := repositories.NewUserRepository(db)
	trackRepo := repositories.NewTrackRepository(db)
	moduleRepo := repositories.NewModuleRepository(db)
	lessonRepo := repositories.NewLessonRepository(db)
	exerciseRepo := repositories.NewExerciseRepository(db)
	progressRepo := repositories.NewProgressRepository(db)

	authSvc := services.NewAuthService(userRepo, r.cfg.JWT.Secret)
	trackSvc := services.NewTrackService(trackRepo)
	moduleSvc := services.NewModuleService(moduleRepo)
	lessonSvc := services.NewLessonService(lessonRepo)
	exerciseSvc := services.NewExerciseService(exerciseRepo)
	progressSvc := services.NewProgressService(progressRepo, lessonRepo, userRepo)

	r.health = controllers.NewHealthController()
	r.auth = controllers.NewAuthController(authSvc)
	r.track = controllers.NewTrackController(trackSvc)
	r.module = controllers.NewModuleController(moduleSvc)
	r.lesson = controllers.NewLessonController(lessonSvc)
	r.exercise = controllers.NewExerciseController(exerciseSvc)
	r.progress = controllers.NewProgressController(progressSvc)
}

func (r *Router) initEngine() {
	gin.SetMode(gin.DebugMode)
	r.engine = gin.New()
	r.engine.Use(gin.Recovery())
	r.engine.Use(middleware.CORSMiddleware())

	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.engine.Group("/api/v1")
	{
		api.GET("/health", r.health.HealthCheck)

		auth := api.Group("/auth")
		{
			auth.POST("/login", r.auth.Login)
			auth.POST("/register", r.auth.Register)
			auth.POST("/logout", r.auth.Logout)
			auth.GET("/me", middleware.AuthMiddleware(), r.auth.Me)
		}

		tracks := api.Group("/tracks")
		{
			tracks.GET("", r.track.GetTracks)
			tracks.GET("/:id", r.track.GetTrack)
			tracks.POST("", middleware.AuthMiddleware(), r.track.CreateTrack)
			tracks.PUT("/:id", middleware.AuthMiddleware(), r.track.UpdateTrack)
			tracks.DELETE("/:id", middleware.AuthMiddleware(), r.track.DeleteTrack)

			tracks.GET("/:id/modules", r.module.GetModulesByTrack)
		}

		modules := api.Group("/modules")
		{
			modules.GET("/:id", r.module.GetModule)
			modules.POST("", middleware.AuthMiddleware(), r.module.CreateModule)
			modules.PUT("/:id", middleware.AuthMiddleware(), r.module.UpdateModule)
			modules.DELETE("/:id", middleware.AuthMiddleware(), r.module.DeleteModule)

			modules.POST("/:id/lessons", middleware.AuthMiddleware(), r.module.LinkLesson)
			modules.GET("/:id/lessons", r.lesson.GetLessonsByModule)
		}

		lessons := api.Group("/lessons")
		{
			lessons.GET("/:id", r.lesson.GetLesson)
			lessons.POST("", middleware.AuthMiddleware(), r.lesson.CreateLesson)
			lessons.PUT("/:id", middleware.AuthMiddleware(), r.lesson.UpdateLesson)
			lessons.DELETE("/:id", middleware.AuthMiddleware(), r.lesson.DeleteLesson)

			lessons.GET("/:id/exercises", r.exercise.GetExercisesByLesson)
		}

		exercises := api.Group("/exercises")
		{
			exercises.GET("/:id", r.exercise.GetExercise)
			exercises.POST("", middleware.AuthMiddleware(), r.exercise.CreateExercise)
			exercises.PUT("/:id", middleware.AuthMiddleware(), r.exercise.UpdateExercise)
			exercises.DELETE("/:id", middleware.AuthMiddleware(), r.exercise.DeleteExercise)
		}

		progress := api.Group("/progress")
		progress.Use(middleware.AuthMiddleware())
		{
			progress.GET("", r.progress.GetMyProgress)
			progress.POST("", r.progress.CompleteLesson)
		}
	}
}

func (r *Router) Run(port string) error {
	return r.engine.Run(":" + port)
}
