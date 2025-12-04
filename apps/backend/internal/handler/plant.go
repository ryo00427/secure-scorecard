package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/model"
)

// GetPlant returns a specific plant
func (h *Handler) GetPlant(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid plant ID",
		})
	}

	plant, err := h.service.GetPlantByID(ctx, uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Plant not found",
		})
	}

	return c.JSON(http.StatusOK, plant)
}

// UpdatePlant updates an existing plant
func (h *Handler) UpdatePlant(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid plant ID",
		})
	}

	plant, err := h.service.GetPlantByID(ctx, uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Plant not found",
		})
	}

	if err := c.Bind(plant); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := h.service.UpdatePlant(ctx, plant); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update plant",
		})
	}

	return c.JSON(http.StatusOK, plant)
}

// DeletePlant deletes a plant
func (h *Handler) DeletePlant(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid plant ID",
		})
	}

	if err := h.service.DeletePlant(ctx, uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete plant",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetPlantCareLogs returns all care logs for a plant
func (h *Handler) GetPlantCareLogs(c echo.Context) error {
	ctx := c.Request().Context()
	plantID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid plant ID",
		})
	}

	careLogs, err := h.service.GetPlantCareLogs(ctx, uint(plantID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch care logs",
		})
	}

	return c.JSON(http.StatusOK, careLogs)
}

// CreateCareLogRequest represents the request body for creating a care log
type CreateCareLogRequest struct {
	Type    string `json:"type" validate:"required"`
	Notes   string `json:"notes"`
	CaredAt string `json:"cared_at"`
}

// CreateCareLog creates a new care log for a plant
func (h *Handler) CreateCareLog(c echo.Context) error {
	ctx := c.Request().Context()
	plantID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid plant ID",
		})
	}

	var req CreateCareLogRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Verify plant exists
	_, err = h.service.GetPlantByID(ctx, uint(plantID))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Plant not found",
		})
	}

	// Parse cared_at time
	caredAt := time.Now()
	if req.CaredAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.CaredAt)
		if err == nil {
			caredAt = parsedTime
		}
	}

	careLog := &model.CareLog{
		PlantID: uint(plantID),
		Type:    req.Type,
		Notes:   req.Notes,
		CaredAt: caredAt,
	}

	if err := h.service.CreateCareLog(ctx, careLog); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create care log",
		})
	}

	return c.JSON(http.StatusCreated, careLog)
}
