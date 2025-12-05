package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the cost parameter for bcrypt hashing
	// Cost 12 provides a good balance between security and performance
	BcryptCost = 12
)

// HashPassword generates a bcrypt hash from a plain text password
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword compares a bcrypt hashed password with a plain text password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
