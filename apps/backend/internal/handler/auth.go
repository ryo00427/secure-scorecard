package handler

import (
	"errors"
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

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name"`
}

// LoginRequest represents the login request body (password-based)
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// FirebaseLoginRequest represents Firebase login request body
type FirebaseLoginRequest struct {
	FirebaseUID string `json:"firebase_uid" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"display_name"`
	PhotoURL    string `json:"photo_url"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// Register handles user registration with email and password
func (h *AuthHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()

	var req RegisterRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return apperrors.NewInternalError("Failed to process registration")
	}

	// Register user
	user, err := h.service.RegisterUser(ctx, req.Email, hashedPassword, req.DisplayName)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			return apperrors.NewConflictError("Email already registered")
		}
		return apperrors.NewInternalError("Failed to register user")
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(user.ID, user.FirebaseUID, user.Email)
	if err != nil {
		return apperrors.NewInternalError("Failed to generate token")
	}

	// Set cookie
	maxAge := int(h.jwtManager.GetExpireDuration().Seconds())
	auth.SetAuthCookie(c, token, maxAge)

	return c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login handles user login with email and password
func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var req LoginRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// Get user by email
	user, err := h.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return apperrors.NewAuthenticationError("Invalid email or password")
	}

	// Check if account is locked
	if h.service.IsAccountLocked(user) {
		return apperrors.NewAuthenticationError("Account is temporarily locked. Please try again later")
	}

	// Verify password
	if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		// Increment failed login count
		_ = h.service.IncrementFailedLogin(ctx, user)
		return apperrors.NewAuthenticationError("Invalid email or password")
	}

	// Reset failed login count on successful login
	if user.FailedLoginCount > 0 {
		_ = h.service.ResetFailedLogin(ctx, user)
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(user.ID, user.FirebaseUID, user.Email)
	if err != nil {
		return apperrors.NewInternalError("Failed to generate token")
	}

	// Set cookie
	maxAge := int(h.jwtManager.GetExpireDuration().Seconds())
	auth.SetAuthCookie(c, token, maxAge)

	return c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// FirebaseLogin handles user login/registration via Firebase
func (h *AuthHandler) FirebaseLogin(c echo.Context) error {
	ctx := c.Request().Context()

	var req FirebaseLoginRequest
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

	return c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract token from header or cookie
	token := c.Request().Header.Get(echo.HeaderAuthorization)
	if token != "" {
		token = auth.ExtractTokenFromHeader(token)
	} else {
		cookie, err := c.Cookie(auth.AuthCookieName)
		if err == nil {
			token = cookie.Value
		}
	}

	// Add token to blacklist if present
	if token != "" {
		tokenHash := auth.HashToken(token)
		expiresAt := h.jwtManager.GetExpireTime()
		if err := h.service.BlacklistToken(ctx, tokenHash, expiresAt); err != nil {
			// Log error but don't fail the logout
			apperrors.NewInternalError("Failed to blacklist token")
		}
	}

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
