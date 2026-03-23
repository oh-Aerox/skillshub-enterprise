package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"skillshub-enterprise/api/internal/config"
	"skillshub-enterprise/api/internal/handlers"
	"skillshub-enterprise/api/internal/middleware"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	dbPool, err := initDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	// Initialize Redis
	rdb := initRedis(cfg.RedisURL)
	defer rdb.Close()

	// Initialize Gin
	gin.SetMode(cfg.GinMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(dbPool, cfg)
	skillHandler := handlers.NewSkillHandler(dbPool, cfg)
	scanHandler := handlers.NewScanHandler(dbPool, cfg)
	reviewHandler := handlers.NewReviewHandler(dbPool, cfg)
	userHandler := handlers.NewUserHandler(dbPool, cfg)

	// Setup routes
	setupRoutes(router, authHandler, skillHandler, scanHandler, reviewHandler, userHandler, cfg)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func initDatabase(databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")
	return pool, nil
}

func initRedis(redisURL string) *redis.Client {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connection established")
	return client
}

func setupRoutes(
	router *gin.Engine,
	authHandler *handlers.AuthHandler,
	skillHandler *handlers.SkillHandler,
	scanHandler *handlers.ScanHandler,
	reviewHandler *handlers.ReviewHandler,
	userHandler *handlers.UserHandler,
	cfg *config.Config,
) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// Public skill search
		v1.GET("/skills", skillHandler.ListSkills)
		v1.GET("/skills/:skillId", skillHandler.GetSkill)
	}

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(dbPool, cfg))
	{
		// User routes
		user := protected.Group("/user")
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
		}

		// Skill routes
		skills := protected.Group("/skills")
		{
			skills.POST("/", skillHandler.CreateSkill)
			skills.PUT("/:skillId", skillHandler.UpdateSkill)
			skills.DELETE("/:skillId", skillHandler.DeleteSkill)
			skills.POST("/:skillId/install", skillHandler.InstallSkill)
			skills.DELETE("/:skillId/install", skillHandler.UninstallSkill)
			skills.GET("/:skillId/versions", skillHandler.ListVersions)
			skills.GET("/:skillId/:version/download", skillHandler.DownloadSkill)
		}

		// Scan routes
		scans := protected.Group("/scans")
		{
			scans.GET("/:scanId", scanHandler.GetScan)
			scans.POST("/trigger", scanHandler.TriggerScan)
		}

		// Review routes
		reviews := protected.Group("/reviews")
		{
			reviews.GET("/", reviewHandler.ListReviews)
			reviews.GET("/:reviewId", reviewHandler.GetReview)
			reviews.PUT("/:reviewId", reviewHandler.DecideReview)
		}

		// Installation routes
		installs := protected.Group("/installations")
		{
			installs.GET("/", skillHandler.ListInstallations)
		}
	}

	// Admin routes
	admin := router.Group("/api/admin/v1")
	admin.Use(middleware.AuthMiddleware(dbPool, cfg))
	admin.Use(middleware.RequireRole("admin", "security"))
	{
		admin.GET("/users", userHandler.ListUsers)
		admin.GET("/stats", skillHandler.GetStats)
		admin.GET("/audit-logs", nil) // TODO: implement audit log handler
		admin.PUT("/scan-rules", nil) // TODO: implement scan rules update
	}
}
