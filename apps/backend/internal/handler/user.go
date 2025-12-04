package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetCurrentUser returns the current authenticated user
func (h *Handler) GetCurrentUser(c echo.Context) error {
	// TODO: Get user from JWT token context
	// For now, return a placeholder response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Authentication not implemented yet",
	})
}
