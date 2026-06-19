package http

import (
	"net/http/httputil"
	"net/url"
	"os"

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
	"github.com/voxlab/voxlab-backend/internal/storage"
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
	user     *controllers.UserController
	upload   *controllers.UploadController
	docs     *controllers.DocsController
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
	lessonExerciseRepo := repositories.NewLessonExerciseRepository(db)
	progressRepo := repositories.NewProgressRepository(db)

	authSvc := services.NewAuthService(userRepo, r.cfg.JWT.Secret)
	trackSvc := services.NewTrackService(trackRepo)
	moduleSvc := services.NewModuleService(moduleRepo)
	lessonSvc := services.NewLessonService(lessonRepo)
	exerciseSvc := services.NewExerciseService(exerciseRepo, lessonExerciseRepo)
	progressSvc := services.NewProgressService(progressRepo, lessonRepo, userRepo)

	userSvc := services.NewUserService(userRepo)
	uploadSvc := services.NewUploadService(
		storage.GetStorage(),
		trackRepo, moduleRepo, lessonRepo, userRepo,
	)

	r.health = controllers.NewHealthController()
	r.auth = controllers.NewAuthController(authSvc)
	r.track = controllers.NewTrackController(trackSvc)
	r.module = controllers.NewModuleController(moduleSvc)
	r.lesson = controllers.NewLessonController(lessonSvc)
	r.exercise = controllers.NewExerciseController(exerciseSvc)
	r.progress = controllers.NewProgressController(progressSvc)
	r.user = controllers.NewUserController(userSvc)
	r.upload = controllers.NewUploadController(uploadSvc)
	r.docs = controllers.NewDocsController()
}

func (r *Router) initEngine() {
	gin.SetMode(gin.DebugMode)
	r.engine = gin.New()
	r.engine.Use(gin.Recovery())
	r.engine.Use(middleware.CORSMiddleware())

	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.engine.GET("/docs/es", r.docs.ServeSwaggerUI)
	r.engine.GET("/api/v1/docs/es/spec", r.docs.ServeTranslatedSpec)

	api := r.engine.Group("/api/v1")
	{
		api.GET("/health", r.health.HealthCheck)

		analyzerTarget := os.Getenv("ANALYZER_URL")
		if analyzerTarget == "" {
			analyzerTarget = "http://localhost:8001"
		}
		analyzerURL, _ := url.Parse(analyzerTarget)
		analyzerProxy := httputil.NewSingleHostReverseProxy(analyzerURL)
		api.GET("/analyzer/openapi.json", func(c *gin.Context) {
			c.Request.URL.Path = "/openapi.json"
			analyzerProxy.ServeHTTP(c.Writer, c.Request)
		})

		auth := api.Group("/auth")
		{
			auth.POST("/login", r.auth.Login)
			auth.POST("/register", r.auth.Register)
			auth.POST("/logout", r.auth.Logout)
			auth.GET("/me", middleware.AuthMiddleware(), r.auth.Me)
			auth.GET("/profile", middleware.AuthMiddleware(), r.auth.GetProfile)
			auth.PUT("/profile", middleware.AuthMiddleware(), r.auth.UpdateProfile)
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

			lessons.POST("/:id/exercises", middleware.AuthMiddleware(), r.exercise.LinkExerciseToLesson)
			lessons.DELETE("/:id/exercises/:exerciseId", middleware.AuthMiddleware(), r.exercise.UnlinkExerciseFromLesson)
			lessons.PUT("/:id/exercises/:exerciseId/reorder", middleware.AuthMiddleware(), r.exercise.ReorderExerciseInLesson)
		}

		exercises := api.Group("/exercises")
		{
			exercises.GET("", r.exercise.ListExercises)
			exercises.GET("/requirement-catalog", r.exercise.GetRequirementCatalog)
			exercises.GET("/:id", r.exercise.GetExercise)
			exercises.POST("", middleware.AuthMiddleware(), r.exercise.CreateExercise)
			exercises.PUT("/:id", middleware.AuthMiddleware(), r.exercise.UpdateExercise)
			exercises.DELETE("/:id", middleware.AuthMiddleware(), r.exercise.DeleteExercise)
			exercises.POST("/analyze-text", middleware.AuthMiddleware(), r.exercise.AnalyzeText)
		}

		progress := api.Group("/progress")
		progress.Use(middleware.AuthMiddleware())
		{
			progress.GET("", r.progress.GetMyProgress)
			progress.POST("", r.progress.CompleteLesson)
		}

		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			users.GET("", r.user.GetUsers)
			users.GET("/:id", r.user.GetUser)
			users.PUT("/:id", r.user.UpdateUser)
			users.DELETE("/:id", r.user.DeleteUser)
		}

		upload := api.Group("/upload")
		{
			upload.POST("/track/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.upload.UploadTrackImage)
			upload.POST("/module/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.upload.UploadModuleImage)
			upload.POST("/lesson/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.upload.UploadLessonImage)
			upload.POST("/avatar", middleware.AuthMiddleware(), r.upload.UploadAvatar)
		}
	}
}

func (r *Router) Run(port string) error {
	return r.engine.Run(":" + port)
}
