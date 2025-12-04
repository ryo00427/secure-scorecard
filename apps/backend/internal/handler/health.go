package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// HelloResponse represents the hello response
type HelloResponse struct {
	Message string `json:"message"`
}

// Health handles the health check endpoint
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Status: "ok",
	})
}

// Hello handles the root endpoint
func (h *Handler) Hello(c echo.Context) error {
	return c.JSON(http.StatusOK, HelloResponse{
		Message: "Welcome to Home Garden Management API",
	})
}
