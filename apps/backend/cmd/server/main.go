package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)
	e.GET("/health", health)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func hello(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Hello, Home Garden Management API!",
	})
}

func health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
