package http

import (
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

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

	health           *controllers.HealthController
	auth             *controllers.AuthController
	track            *controllers.TrackController
	module           *controllers.ModuleController
	lesson           *controllers.LessonController
	exercise         *controllers.ExerciseController
	progress         *controllers.ProgressController
	exerciseProgress *controllers.ExerciseProgressController
	reaction         *controllers.ReactionController
	user             *controllers.UserController
	upload           *controllers.UploadController
	docs             *controllers.DocsController
	learningPath     *controllers.LearningPathController
	clientLog        *controllers.ClientLogController
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
	attemptRepo := repositories.NewExerciseAttemptRepository(db)
	exerciseProgressRepo := repositories.NewExerciseProgressRepository(db)

	lifeSvc := services.NewLifeService(userRepo)
	streakSvc := services.NewStreakService(userRepo)
	authSvc := services.NewAuthService(userRepo, r.cfg.JWT.Secret)
	trackSvc := services.NewTrackService(trackRepo)
	moduleSvc := services.NewModuleService(moduleRepo)
	lessonSvc := services.NewLessonService(lessonRepo)
	exerciseSvc := services.NewExerciseService(exerciseRepo, lessonExerciseRepo)
	progressSvc := services.NewProgressService(progressRepo, lessonRepo, userRepo, lifeSvc, streakSvc, attemptRepo)
	attemptSvc := services.NewAttemptService(lifeSvc, streakSvc, exerciseRepo, userRepo, attemptRepo, progressRepo)
	exerciseProgressSvc := services.NewExerciseProgressService(exerciseProgressRepo, userRepo, exerciseRepo)

	clientLogRepo := repositories.NewClientLogRepository(db)
	clientLogSvc := services.NewClientLogService(clientLogRepo)

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
	r.exercise = controllers.NewExerciseController(exerciseSvc, attemptSvc)
	r.progress = controllers.NewProgressController(progressSvc)
	r.exerciseProgress = controllers.NewExerciseProgressController(exerciseProgressSvc)
	r.user = controllers.NewUserController(userSvc, lifeSvc, streakSvc, db)
	r.upload = controllers.NewUploadController(uploadSvc)
	r.docs = controllers.NewDocsController()
	r.learningPath = controllers.NewLearningPathController(trackRepo, moduleRepo, lessonRepo, progressRepo, lifeSvc, streakSvc)
	r.clientLog = controllers.NewClientLogController(clientLogSvc, clientLogRepo)
}

func (r *Router) initEngine() {
	gin.SetMode(ginMode(r.cfg.AppEnv))

	r.engine = gin.New()

	// El orden importa: RequestID debe correr antes que RequestLogger para que
	// cada línea del access log pueda incluir el ID de correlación.
	r.engine.Use(middleware.RequestID())
	r.engine.Use(middleware.RequestLogger())
	r.engine.Use(gin.RecoveryWithWriter(os.Stdout))
	r.engine.Use(middleware.CORSMiddleware())

	r.engine.GET("/api/v1/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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
			auth.DELETE("/me", middleware.AuthMiddleware(), r.auth.DeleteAccount)
			auth.GET("/profile", middleware.AuthMiddleware(), r.auth.GetProfile)
			auth.PUT("/profile", middleware.AuthMiddleware(), r.auth.UpdateProfile)
		}

		// Contenido: lectura pública; crear, editar, borrar y reordenar solo
		// admin (antes bastaba cualquier token de la app).
		tracks := api.Group("/tracks")
		{
			tracks.GET("", r.track.GetTracks)
			tracks.GET("/:id", r.track.GetTrack)
			tracks.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.track.CreateTrack)
			tracks.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.track.UpdateTrack)
			tracks.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.track.DeleteTrack)
			tracks.PUT("/reorder", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.track.ReorderTracks)

			tracks.GET("/:id/modules", r.module.GetModulesByTrack)
			tracks.PUT("/:id/modules/reorder", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.module.ReorderModules)
		}

		modules := api.Group("/modules")
		{
			modules.GET("/:id", r.module.GetModule)
			modules.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.module.CreateModule)
			modules.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.module.UpdateModule)
			modules.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.module.DeleteModule)

			modules.POST("/:id/lessons", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.module.LinkLesson)
			modules.GET("/:id/lessons", r.lesson.GetLessonsByModule)
			modules.DELETE("/:id/lessons/:lessonId", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.module.UnlinkLesson)
			modules.PUT("/:id/lessons/reorder", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.module.ReorderLessons)
		}

		api.GET("/lessons/shared", r.lesson.GetSharedLessons)

		lessons := api.Group("/lessons")
		{
			lessons.GET("", r.lesson.ListLessons)
			lessons.GET("/:id", r.lesson.GetLesson)
			lessons.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.lesson.CreateLesson)
			lessons.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.lesson.UpdateLesson)
			lessons.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.lesson.DeleteLesson)

			lessons.GET("/:id/modules", r.lesson.GetModulesByLesson)
			lessons.GET("/:id/exercises", r.lesson.GetExercisesByLesson)
			lessons.POST("/:id/exercises", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.exercise.LinkExerciseToLesson)
			lessons.DELETE("/:id/exercises/:exerciseId", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.exercise.UnlinkExerciseFromLesson)
			lessons.PUT("/:id/exercises/reorder", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.exercise.BatchReorderExercises)
			lessons.PUT("/:id/exercises/:exerciseId/reorder", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.exercise.ReorderExerciseInLesson)
		}

		exercises := api.Group("/exercises")
		{
			exercises.GET("", r.exercise.ListExercises)
			exercises.GET("/requirement-catalog", r.exercise.GetRequirementCatalog)
			exercises.GET("/:id", r.exercise.GetExercise)
			exercises.GET("/:id/lessons", r.exercise.GetLessonsByExercise)
			exercises.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.exercise.CreateExercise)
			exercises.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.exercise.UpdateExercise)
			exercises.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.exercise.DeleteExercise)
			exercises.POST("/analyze-text", middleware.AuthMiddleware(), r.exercise.AnalyzeText)
			exercises.POST("/analyze-speech", middleware.AuthMiddleware(), r.exercise.AnalyzeSpeech)
			exercises.POST("/:id/attempt", middleware.AuthMiddleware(), r.exercise.AttemptExercise)
		}

		learningPath := api.Group("/learning-path")
		learningPath.Use(middleware.AuthMiddleware())
		{
			learningPath.GET("", r.learningPath.GetLearningPath)
		}

		api.GET("/users/lives", middleware.AuthMiddleware(), r.user.GetLives)
		api.POST("/users/streak/recover", middleware.AuthMiddleware(), r.user.RecoverStreak)
		api.GET("/leaderboard", middleware.AuthMiddleware(), r.user.GetLeaderboard)

		progress := api.Group("/lesson-progress")
		progress.Use(middleware.AuthMiddleware())
		{
			progress.GET("", r.progress.GetMyProgress)
			progress.GET("/weekly", r.progress.GetWeeklyProgress)
			progress.POST("/sync", r.progress.SyncProgress)
			progress.POST("", r.progress.CompleteLesson)
			progress.PATCH("/:lesson_id", r.progress.UpdateProgress)
		}

		exerciseProgress := api.Group("/progress")
		exerciseProgress.Use(middleware.AuthMiddleware())
		{
			exerciseProgress.GET("", r.exerciseProgress.GetProgress)
			exerciseProgress.POST("", r.exerciseProgress.CompleteExercise)
			exerciseProgress.POST("/sync", r.exerciseProgress.SyncProgress)
		}

		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			users.GET("", r.user.GetUsers)
			users.GET("/:id", r.user.GetUser)
			users.PUT("/:id", r.user.UpdateUser)
			users.DELETE("/:id", r.user.DeleteUser)
		}

		// Logs del cliente. POST requiere token (cualquier admin logueado);
		// GET exige admin porque expone URLs y stacks de otros usuarios.
		logs := api.Group("/logs")
		{
			logs.POST("", middleware.AuthMiddleware(), r.clientLog.Ingest)
			logs.GET("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), r.clientLog.ListRecent)
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

// ginMode mapea APP_ENV al modo de gin. Antes estaba hardcodeado a DebugMode,
// lo que en producción activaba warnings de routing y verbosity innecesaria.
func ginMode(appEnv string) string {
	if strings.EqualFold(appEnv, "production") {
		return gin.ReleaseMode
	}
	return gin.DebugMode
}

func (r *Router) Run(port string) error {
	return r.engine.Run(":" + port)
}
