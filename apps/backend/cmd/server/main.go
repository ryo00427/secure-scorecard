package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/auth"
	"github.com/secure-scorecard/backend/internal/config"
	"github.com/secure-scorecard/backend/internal/database"
	apperrors "github.com/secure-scorecard/backend/internal/errors"
	"github.com/secure-scorecard/backend/internal/handler"
	"github.com/secure-scorecard/backend/internal/middleware"
	"github.com/secure-scorecard/backend/internal/repository"
	"github.com/secure-scorecard/backend/internal/service"
	"github.com/secure-scorecard/backend/internal/storage"
	"github.com/secure-scorecard/backend/internal/validator"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup structured logging
	setupLogging(cfg)

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Set custom error handler
	e.HTTPErrorHandler = apperrors.ErrorHandler

	// Set custom validator
	e.Validator = validator.NewValidator()

	// Setup middleware
	middleware.SetupMiddleware(e, cfg)

	// Initialize database
	db, err := database.Connect(cfg, nil)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Println("Running in standalone mode without database")
		setupStandaloneRoutes(e)
	} else {
		defer db.Close()

		// Run full database setup (migrations, indexes, constraints, materialized views)
		if err := db.Setup(); err != nil {
			log.Printf("Warning: Database setup failed: %v", err)
		}

		// Initialize JWT manager
		jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireHour)

		// Initialize S3 service (optional - can run without S3)
		s3Config := &storage.S3Config{
			Region:          cfg.S3.Region,
			BucketName:      cfg.S3.BucketName,
			AccessKeyID:     cfg.S3.AccessKeyID,
			SecretAccessKey: cfg.S3.SecretAccessKey,
			CloudFrontURL:   cfg.S3.CloudFrontURL,
			Endpoint:        cfg.S3.Endpoint,
		}
		s3Svc, err := storage.NewS3Service(s3Config)
		if err != nil {
			log.Printf("Warning: S3 service initialization failed: %v", err)
			log.Println("Image upload functionality will be unavailable")
			s3Svc = nil
		} else if !s3Config.IsConfigured() {
			log.Println("S3 not configured - image upload functionality will be unavailable")
		}

		// Initialize layers with new repository manager
		repos := repository.NewRepositoryManager(db.DB)
		svc := service.NewService(repos)
		h := handler.NewHandler(svc, jwtManager, s3Svc)

		// Register routes
		h.RegisterRoutes(e)

		// Register scheduler routes (for EventBridge Scheduler)
		h.RegisterSchedulerRoutes(e, cfg.Scheduler.AuthToken)

		// Add database health check endpoint
		e.GET("/health/db", func(c echo.Context) error {
			if err := db.HealthCheck(); err != nil {
				return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				})
			}
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status": "healthy",
				"stats":  db.Stats(),
			})
		})
	}

	// Start server with graceful shutdown
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Printf("Starting server on %s (env: %s)", addr, cfg.Server.Env)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown with timeout
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// setupLogging configures structured logging
func setupLogging(cfg *config.Config) {
	var level slog.Level
	switch cfg.Server.Env {
	case "production":
		level = slog.LevelInfo
	case "development":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Logging initialized", "env", cfg.Server.Env, "level", level.String())
}

// startTokenCleanupJob starts a background job to clean up expired tokens
func startTokenCleanupJob(svc *service.Service) {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	defer ticker.Stop()

	// Run immediately on startup
	cleanupExpiredTokens(svc)

	// Then run daily
	for range ticker.C {
		cleanupExpiredTokens(svc)
	}
}

// cleanupExpiredTokens removes expired tokens from the blacklist
func cleanupExpiredTokens(svc *service.Service) {
	ctx := context.Background()
	if err := svc.CleanupExpiredTokens(ctx); err != nil {
		slog.Error("Failed to cleanup expired tokens", "error", err)
	} else {
		slog.Info("Expired tokens cleaned up successfully")
	}
}

// setupStandaloneRoutes sets up routes for standalone mode (without database)
func setupStandaloneRoutes(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Welcome to Home Garden Management API",
			"mode":    "standalone",
		})
	})

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"mode":   "standalone",
		})
	})
}
