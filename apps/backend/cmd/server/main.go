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
	"github.com/secure-scorecard/backend/internal/database"
	"github.com/secure-scorecard/backend/internal/handler"
	"github.com/secure-scorecard/backend/internal/middleware"
	"github.com/secure-scorecard/backend/internal/repository"
	"github.com/secure-scorecard/backend/internal/service"
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

	// Initialize database
	db, err := database.Connect(cfg, nil)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Println("Running in standalone mode without database")
		setupStandaloneRoutes(e)
	} else {
		defer db.Close()

		// Run migrations
		if err := db.AutoMigrate(); err != nil {
			log.Printf("Warning: Auto-migration failed: %v", err)
		}

		// Create indexes
		if err := db.CreateIndexes(); err != nil {
			log.Printf("Warning: Index creation failed: %v", err)
		}

		// Initialize layers with new repository manager
		repos := repository.NewRepositoryManager(db.DB)
		svc := service.NewService(repos)
		h := handler.NewHandler(svc)

		// Register routes
		h.RegisterRoutes(e)

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
