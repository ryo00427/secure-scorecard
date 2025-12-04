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

	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/config"
	"github.com/secure-scorecard/backend/internal/handler"
	"github.com/secure-scorecard/backend/internal/middleware"
	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
	"github.com/secure-scorecard/backend/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Setup middleware
	middleware.SetupMiddleware(e)

	// Initialize database (optional - skip if DB not available)
	db, err := initDatabase(cfg)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Println("Running in standalone mode without database")
		// Create handler without service for standalone mode
		setupStandaloneRoutes(e)
	} else {
		// Auto-migrate models
		if err := autoMigrate(db); err != nil {
			log.Printf("Warning: Auto-migration failed: %v", err)
		}

		// Initialize layers
		repo := repository.NewRepository(db)
		svc := service.NewService(repo)
		h := handler.NewHandler(svc)

		// Register routes
		h.RegisterRoutes(e)
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

// initDatabase initializes the database connection
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	if cfg.Server.Env == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// autoMigrate runs database migrations
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Garden{},
		&model.Plant{},
		&model.CareLog{},
	)
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
