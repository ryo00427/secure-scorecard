package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// GetPlant returns a specific plant
func (h *Handler) GetPlant(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid plant ID",
		})
	}

	plant, err := h.service.GetPlantByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Plant not found",
		})
	}

	return c.JSON(http.StatusOK, plant)
}

// UpdatePlant updates an existing plant
func (h *Handler) UpdatePlant(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid plant ID",
		})
	}

	plant, err := h.service.GetPlantByID(uint(id))
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

	if err := h.service.UpdatePlant(plant); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update plant",
		})
	}

	return c.JSON(http.StatusOK, plant)
}

// DeletePlant deletes a plant
func (h *Handler) DeletePlant(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid plant ID",
		})
	}

	if err := h.service.DeletePlant(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete plant",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetPlantCareLogs returns all care logs for a plant
func (h *Handler) GetPlantCareLogs(c echo.Context) error {
	plantID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid plant ID",
		})
	}

	careLogs, err := h.service.GetPlantCareLogs(uint(plantID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch care logs",
		})
	}

	return c.JSON(http.StatusOK, careLogs)
}

// CreateCareLog creates a new care log for a plant
func (h *Handler) CreateCareLog(c echo.Context) error {
	plantID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid plant ID",
		})
	}

	var req struct {
		Type    string `json:"type"`
		Notes   string `json:"notes"`
		CaredAt string `json:"cared_at"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Verify plant exists
	_, err = h.service.GetPlantByID(uint(plantID))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Plant not found",
		})
	}

	// TODO: Parse cared_at time and create care log
	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Care log created",
	})
}
