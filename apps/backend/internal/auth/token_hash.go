package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken creates a SHA-256 hash of the token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ExtractTokenFromHeader extracts the token from Authorization header
// Returns empty string if header is invalid or empty
func ExtractTokenFromHeader(authHeader string) string {
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) {
		return ""
	}
	if authHeader[:len(bearerPrefix)] != bearerPrefix {
		return ""
	}
	return authHeader[len(bearerPrefix):]
}
