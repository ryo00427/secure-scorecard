package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/auth"
	apperrors "github.com/secure-scorecard/backend/internal/errors"
	"github.com/secure-scorecard/backend/internal/service"
	"github.com/secure-scorecard/backend/internal/validator"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	service    *service.Service
	jwtManager *auth.JWTManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(svc *service.Service, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		service:    svc,
		jwtManager: jwtManager,
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	FirebaseUID string `json:"firebase_uid" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"display_name"`
	PhotoURL    string `json:"photo_url"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// Login handles user login/registration
func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var req LoginRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// Get or create user
	user, err := h.service.GetOrCreateUser(ctx, req.FirebaseUID, req.Email, req.DisplayName, req.PhotoURL)
	if err != nil {
		return apperrors.NewInternalError("Failed to process login")
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(user.ID, user.FirebaseUID, user.Email)
	if err != nil {
		return apperrors.NewInternalError("Failed to generate token")
	}

	// Set cookie
	maxAge := int(h.jwtManager.GetExpireDuration().Seconds())
	auth.SetAuthCookie(c, token, maxAge)

	return c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	auth.ClearAuthCookie(c)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	claims := auth.GetUserFromContext(c)
	if claims == nil {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// Generate new token
	token, err := h.jwtManager.GenerateToken(claims.UserID, claims.FirebaseUID, claims.Email)
	if err != nil {
		return apperrors.NewInternalError("Failed to refresh token")
	}

	// Update cookie
	maxAge := int(h.jwtManager.GetExpireDuration().Seconds())
	auth.SetAuthCookie(c, token, maxAge)

	return c.JSON(http.StatusOK, map[string]string{
		"token": token,
	})
}

// Me returns the current user info
func (h *AuthHandler) Me(c echo.Context) error {
	ctx := c.Request().Context()
	claims := auth.GetUserFromContext(c)
	if claims == nil {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	user, err := h.service.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return apperrors.NewNotFoundError("User")
	}

	return c.JSON(http.StatusOK, user)
}
