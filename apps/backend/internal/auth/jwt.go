package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents the JWT claims
type Claims struct {
	UserID      uint   `json:"user_id"`
	FirebaseUID string `json:"firebase_uid"`
	Email       string `json:"email"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey  []byte
	expireHour int
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string, expireHour int) *JWTManager {
	return &JWTManager{
		secretKey:  []byte(secret),
		expireHour: expireHour,
	}
}

// GenerateToken generates a new JWT token for a user
func (m *JWTManager) GenerateToken(userID uint, firebaseUID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:      userID,
		FirebaseUID: firebaseUID,
		Email:       email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(m.expireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "home-garden-api",
			Subject:   firebaseUID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken generates a new token if the current one is still valid
func (m *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return m.GenerateToken(claims.UserID, claims.FirebaseUID, claims.Email)
}

// GetExpireDuration returns the token expiration duration
func (m *JWTManager) GetExpireDuration() time.Duration {
	return time.Duration(m.expireHour) * time.Hour
}
