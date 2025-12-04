package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/service"
)

// Handler holds all HTTP handlers
type Handler struct {
	service *service.Service
}

// NewHandler creates a new Handler instance
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes registers all routes
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	// Health check
	e.GET("/health", h.Health)
	e.GET("/", h.Hello)

	// API v1 group
	api := e.Group("/api/v1")

	// Gardens endpoints
	gardens := api.Group("/gardens")
	gardens.GET("", h.GetGardens)
	gardens.POST("", h.CreateGarden)
	gardens.GET("/:id", h.GetGarden)
	gardens.PUT("/:id", h.UpdateGarden)
	gardens.DELETE("/:id", h.DeleteGarden)

	// Plants endpoints (nested under gardens)
	gardens.GET("/:id/plants", h.GetGardenPlants)
	gardens.POST("/:id/plants", h.CreatePlant)

	// Plants endpoints (direct access)
	plants := api.Group("/plants")
	plants.GET("/:id", h.GetPlant)
	plants.PUT("/:id", h.UpdatePlant)
	plants.DELETE("/:id", h.DeletePlant)

	// Care logs endpoints (nested under plants)
	plants.GET("/:id/care-logs", h.GetPlantCareLogs)
	plants.POST("/:id/care-logs", h.CreateCareLog)

	// User endpoints
	users := api.Group("/users")
	users.GET("/me", h.GetCurrentUser)
}
