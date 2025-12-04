package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	// AuthCookieName is the name of the authentication cookie
	AuthCookieName = "auth_token"
	// UserContextKey is the key used to store user claims in context
	UserContextKey = "user"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(jwtManager *JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractToken(c)
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authentication token")
			}

			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				if err == ErrExpiredToken {
					return echo.NewHTTPError(http.StatusUnauthorized, "token has expired")
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			// Store claims in context
			c.Set(UserContextKey, claims)

			return next(c)
		}
	}
}

// OptionalAuthMiddleware creates a middleware that extracts user if token is present
// but doesn't require authentication
func OptionalAuthMiddleware(jwtManager *JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := extractToken(c)
			if token != "" {
				claims, err := jwtManager.ValidateToken(token)
				if err == nil {
					c.Set(UserContextKey, claims)
				}
			}
			return next(c)
		}
	}
}

// extractToken extracts the JWT token from the request
// It checks the Authorization header first, then falls back to cookies
func extractToken(c echo.Context) string {
	// Check Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Check cookie
	cookie, err := c.Cookie(AuthCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// GetUserFromContext retrieves the user claims from the context
func GetUserFromContext(c echo.Context) *Claims {
	user := c.Get(UserContextKey)
	if user == nil {
		return nil
	}
	claims, ok := user.(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(c echo.Context) uint {
	claims := GetUserFromContext(c)
	if claims == nil {
		return 0
	}
	return claims.UserID
}

// SetAuthCookie sets the authentication cookie
func SetAuthCookie(c echo.Context, token string, maxAge int) {
	cookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   c.Scheme() == "https",
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}

// ClearAuthCookie clears the authentication cookie
func ClearAuthCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.Scheme() == "https",
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}
