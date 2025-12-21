package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/yourusername/rankedterview-backend/internal/config"
	"github.com/yourusername/rankedterview-backend/internal/database"
	"github.com/yourusername/rankedterview-backend/internal/handlers"
	"github.com/yourusername/rankedterview-backend/internal/middleware"
	"github.com/yourusername/rankedterview-backend/internal/repositories"
	"github.com/yourusername/rankedterview-backend/internal/services"
	"github.com/yourusername/rankedterview-backend/internal/websocket"
	"github.com/yourusername/rankedterview-backend/pkg/logger"
)

func main() {
	// Load environment variables - try parent directory first (root .env), then current directory
	if err := godotenv.Load("../.env"); err != nil {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	// Initialize configuration
	cfg := config.LoadConfig()

	// Initialize logger
	loggerInstance := logger.NewLogger(cfg.Environment)

	// Initialize database connections
	loggerInstance.Info("Connecting to MongoDB...")
	mongoDB, err := database.NewMongoDB(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		loggerInstance.Fatal("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Disconnect()

	loggerInstance.Info("Connecting to Redis...")
	redisClient := database.NewRedis(cfg.RedisURI, cfg.RedisPassword, cfg.RedisDB)
	defer redisClient.Close()

	// Ping databases
	if err := mongoDB.Ping(context.Background()); err != nil {
		loggerInstance.Fatal("MongoDB ping failed: %v", err)
	}
	loggerInstance.Info("MongoDB connected successfully")

	if err := redisClient.Ping(context.Background()); err != nil {
		loggerInstance.Fatal("Redis ping failed: %v", err)
	}
	loggerInstance.Info("Redis connected successfully")

	// Initialize repositories
	userRepo := repositories.NewUserRepository(mongoDB)
	interviewRepo := repositories.NewInterviewRepository(mongoDB)
	roomRepo := repositories.NewRoomRepository(mongoDB)
	rankingRepo := repositories.NewRankingRepository(mongoDB)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	userService := services.NewUserService(userRepo)
	matchmakingService := services.NewMatchmakingService(redisClient, roomRepo)
	roomService := services.NewRoomService(roomRepo, redisClient)
	interviewService := services.NewInterviewService(interviewRepo, roomRepo)
	rankingService := services.NewRankingService(rankingRepo, redisClient)

	// Initialize WebSocket hub
	hub := websocket.NewHub(redisClient)
	go hub.Run()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	matchmakingHandler := handlers.NewMatchmakingHandler(matchmakingService, hub)
	roomHandler := handlers.NewRoomHandler(roomService)
	interviewHandler := handlers.NewInterviewHandler(interviewService)
	rankingHandler := handlers.NewRankingHandler(rankingService)
	webhookHandler := handlers.NewWebhookHandler(interviewService, rankingService, cfg)
	wsHandler := handlers.NewWebSocketHandler(hub)

	// Set up Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(loggerInstance))
	router.Use(middleware.CORS(cfg.AllowedOrigins))
	router.Use(middleware.RateLimiter(redisClient))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/oauth/google", authHandler.GoogleOAuth)
			auth.GET("/oauth/github", authHandler.GitHubOAuth)
			auth.POST("/callback", authHandler.OAuthCallback)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", userHandler.GetCurrentUser)
				users.PUT("/me", userHandler.UpdateProfile)
				users.GET("/:id", userHandler.GetUser)
				users.GET("/:id/stats", userHandler.GetUserStats)
			}

			// Matchmaking routes
			matchmaking := protected.Group("/matchmaking")
			{
				matchmaking.POST("/join", matchmakingHandler.JoinQueue)
				matchmaking.POST("/leave", matchmakingHandler.LeaveQueue)
				matchmaking.GET("/status", matchmakingHandler.GetQueueStatus)
			}

			// Room routes
			rooms := protected.Group("/rooms")
			{
				rooms.GET("/:roomId", roomHandler.GetRoom)
				rooms.POST("/:roomId/join", roomHandler.JoinRoom)
				rooms.POST("/:roomId/leave", roomHandler.LeaveRoom)
				rooms.GET("/:roomId/state", roomHandler.GetRoomState)
			}

			// Interview routes
			interviews := protected.Group("/interviews")
			{
				interviews.GET("", interviewHandler.ListInterviews)
				interviews.GET("/:id", interviewHandler.GetInterview)
				interviews.GET("/:id/transcript", interviewHandler.GetTranscript)
				interviews.GET("/:id/recording", interviewHandler.GetRecordingURLs)
				interviews.GET("/:id/feedback", interviewHandler.GetFeedback)
			}

			// Ranking routes
			rankings := protected.Group("/rankings")
			{
				rankings.GET("/global", rankingHandler.GetGlobalLeaderboard)
				rankings.GET("/category/:category", rankingHandler.GetCategoryLeaderboard)
				rankings.GET("/user/:userId", rankingHandler.GetUserRank)
				rankings.GET("/history/:userId", rankingHandler.GetRankHistory)
			}
		}

		// Webhook routes (authenticated differently)
		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("/recall", webhookHandler.RecallWebhook)
		}
	}

	// WebSocket route
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		loggerInstance.Info("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggerInstance.Fatal("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	loggerInstance.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		loggerInstance.Fatal("Server forced to shutdown: %v", err)
	}

	loggerInstance.Info("Server exited")
}
