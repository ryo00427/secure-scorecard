package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/model"
)

// CreateGardenRequest represents the request body for creating a garden
type CreateGardenRequest struct {
	Name        string  `json:"name" validate:"required,max=100"`
	Description string  `json:"description" validate:"max=500"`
	Location    string  `json:"location" validate:"max=200"`
	SizeM2      float64 `json:"size_m2" validate:"gte=0"`
}

// UpdateGardenRequest represents the request body for updating a garden
type UpdateGardenRequest struct {
	Name        string  `json:"name" validate:"max=100"`
	Description string  `json:"description" validate:"max=500"`
	Location    string  `json:"location" validate:"max=200"`
	SizeM2      float64 `json:"size_m2" validate:"gte=0"`
}

// GetGardens returns all gardens for the current user
func (h *Handler) GetGardens(c echo.Context) error {
	ctx := c.Request().Context()
	// TODO: Get user ID from JWT token
	userID := uint(1) // Placeholder

	gardens, err := h.service.GetUserGardens(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch gardens",
		})
	}

	return c.JSON(http.StatusOK, gardens)
}

// GetGarden returns a specific garden
func (h *Handler) GetGarden(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid garden ID",
		})
	}

	garden, err := h.service.GetGardenByID(ctx, uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Garden not found",
		})
	}

	return c.JSON(http.StatusOK, garden)
}

// CreateGarden creates a new garden
func (h *Handler) CreateGarden(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateGardenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// TODO: Get user ID from JWT token
	userID := uint(1) // Placeholder

	garden, err := h.service.CreateGarden(ctx, userID, req.Name, req.Description, req.Location, req.SizeM2)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create garden",
		})
	}

	return c.JSON(http.StatusCreated, garden)
}

// UpdateGarden updates an existing garden
func (h *Handler) UpdateGarden(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid garden ID",
		})
	}

	var req UpdateGardenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	garden, err := h.service.GetGardenByID(ctx, uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Garden not found",
		})
	}

	// Update fields
	if req.Name != "" {
		garden.Name = req.Name
	}
	if req.Description != "" {
		garden.Description = req.Description
	}
	if req.Location != "" {
		garden.Location = req.Location
	}
	if req.SizeM2 > 0 {
		garden.SizeM2 = req.SizeM2
	}

	if err := h.service.UpdateGarden(ctx, garden); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update garden",
		})
	}

	return c.JSON(http.StatusOK, garden)
}

// DeleteGarden deletes a garden
func (h *Handler) DeleteGarden(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid garden ID",
		})
	}

	if err := h.service.DeleteGarden(ctx, uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete garden",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetGardenPlants returns all plants in a garden
func (h *Handler) GetGardenPlants(c echo.Context) error {
	ctx := c.Request().Context()
	gardenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid garden ID",
		})
	}

	plants, err := h.service.GetGardenPlants(ctx, uint(gardenID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch plants",
		})
	}

	return c.JSON(http.StatusOK, plants)
}

// CreatePlant creates a new plant in a garden
func (h *Handler) CreatePlant(c echo.Context) error {
	ctx := c.Request().Context()
	gardenID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid garden ID",
		})
	}

	var plant model.Plant
	if err := c.Bind(&plant); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	plant.GardenID = uint(gardenID)

	if err := h.service.CreatePlant(ctx, &plant); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create plant",
		})
	}

	return c.JSON(http.StatusCreated, plant)
}
