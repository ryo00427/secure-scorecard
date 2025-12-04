package errors

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorHandler is a custom error handler for Echo
func ErrorHandler(err error, c echo.Context) {
	// Default error
	code := http.StatusInternalServerError
	message := "Internal server error"
	errorCode := ErrCodeInternal
	var details any

	// Handle AppError
	if appErr, ok := err.(*AppError); ok {
		code = appErr.StatusCode
		message = appErr.Message
		errorCode = appErr.Code
		details = appErr.Details
	} else if echoErr, ok := err.(*echo.HTTPError); ok {
		// Handle Echo's HTTPError
		code = echoErr.Code
		if msg, ok := echoErr.Message.(string); ok {
			message = msg
		}
		// Map Echo errors to our error codes
		switch code {
		case http.StatusBadRequest:
			errorCode = ErrCodeBadRequest
		case http.StatusUnauthorized:
			errorCode = ErrCodeAuthentication
		case http.StatusForbidden:
			errorCode = ErrCodeAuthorization
		case http.StatusNotFound:
			errorCode = ErrCodeNotFound
		case http.StatusConflict:
			errorCode = ErrCodeConflict
		default:
			errorCode = ErrCodeInternal
		}
	}

	// Log error
	logError(c, err, code)

	// Don't send response if already committed
	if c.Response().Committed {
		return
	}

	// Send JSON response
	response := ErrorResponse{
		Error: AppError{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}

	if err := c.JSON(code, response); err != nil {
		slog.Error("Failed to send error response", "error", err)
	}
}

// logError logs the error with appropriate level
func logError(c echo.Context, err error, statusCode int) {
	attrs := []any{
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"status", statusCode,
		"error", err.Error(),
	}

	// Add user ID if available
	if userID := c.Get("user"); userID != nil {
		attrs = append(attrs, "user_id", userID)
	}

	// Add request ID if available
	if reqID := c.Response().Header().Get(echo.HeaderXRequestID); reqID != "" {
		attrs = append(attrs, "request_id", reqID)
	}

	// Log with appropriate level
	if statusCode >= 500 {
		slog.Error("HTTP request error", attrs...)
	} else if statusCode >= 400 {
		slog.Warn("HTTP request warning", attrs...)
	} else {
		slog.Info("HTTP request", attrs...)
	}
}
