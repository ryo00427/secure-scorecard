package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/secure-scorecard/backend/internal/config"
)

// SetupMiddleware configures all middleware for the application
func SetupMiddleware(e *echo.Echo, cfg *config.Config) {
	// Request ID
	e.Use(middleware.RequestID())

	// Logger
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${id} ${method} ${uri} ${status} ${latency_human}\n",
	}))

	// Recover from panics
	e.Use(middleware.Recover())

	// CORS with whitelisted origins
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true, // Required for cookies
	}))

	// Secure headers
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))
}

