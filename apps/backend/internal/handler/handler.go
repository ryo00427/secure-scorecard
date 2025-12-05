package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/auth"
	"github.com/secure-scorecard/backend/internal/service"
)

// Handler holds all HTTP handlers
type Handler struct {
	service    *service.Service
	jwtManager *auth.JWTManager
}

// NewHandler creates a new Handler instance
func NewHandler(svc *service.Service, jwtManager *auth.JWTManager) *Handler {
	return &Handler{
		service:    svc,
		jwtManager: jwtManager,
	}
}

// RegisterRoutes registers all routes
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	// Health check (public)
	e.GET("/health", h.Health)
	e.GET("/", h.Hello)

	// API v1 group
	api := e.Group("/api/v1")

	// Auth endpoints (public)
	authHandler := NewAuthHandler(h.service, h.jwtManager)
	authGroup := api.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/firebase-login", authHandler.FirebaseLogin)
	authGroup.POST("/logout", authHandler.Logout)

	// Protected auth endpoints
	authProtected := authGroup.Group("")
	authProtected.Use(auth.AuthMiddleware(h.jwtManager, h.service))
	authProtected.POST("/refresh", authHandler.RefreshToken)
	authProtected.GET("/me", authHandler.Me)

	// Protected API endpoints
	protected := api.Group("")
	protected.Use(auth.AuthMiddleware(h.jwtManager, h.service))

	// Gardens endpoints (protected)
	gardens := protected.Group("/gardens")
	gardens.GET("", h.GetGardens)
	gardens.POST("", h.CreateGarden)
	gardens.GET("/:id", h.GetGarden)
	gardens.PUT("/:id", h.UpdateGarden)
	gardens.DELETE("/:id", h.DeleteGarden)

	// Plants endpoints (nested under gardens, protected)
	gardens.GET("/:id/plants", h.GetGardenPlants)
	gardens.POST("/:id/plants", h.CreatePlant)

	// Plants endpoints (direct access, protected)
	plants := protected.Group("/plants")
	plants.GET("/:id", h.GetPlant)
	plants.PUT("/:id", h.UpdatePlant)
	plants.DELETE("/:id", h.DeletePlant)

	// Care logs endpoints (nested under plants, protected)
	plants.GET("/:id/care-logs", h.GetPlantCareLogs)
	plants.POST("/:id/care-logs", h.CreateCareLog)

	// User endpoints (protected)
	users := protected.Group("/users")
	users.GET("/me", h.GetCurrentUser)
}
