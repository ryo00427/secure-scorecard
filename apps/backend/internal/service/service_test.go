package service

import (
	"context"
	"testing"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
)

// TestRegisterUser_Success tests successful user registration
func TestRegisterUser_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	user, err := svc.RegisterUser(ctx, "test@example.com", "hashedpassword", "Test User")
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	if user.DisplayName != "Test User" {
		t.Errorf("Expected display name 'Test User', got '%s'", user.DisplayName)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}
}

// TestRegisterUser_DuplicateEmail tests registration with existing email
func TestRegisterUser_DuplicateEmail(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// Create first user
	_, err := svc.RegisterUser(ctx, "duplicate@example.com", "hashedpassword", "First User")
	if err != nil {
		t.Fatalf("First RegisterUser failed: %v", err)
	}

	// Try to create second user with same email
	_, err = svc.RegisterUser(ctx, "duplicate@example.com", "hashedpassword", "Second User")
	if err == nil {
		t.Error("Expected error for duplicate email")
	}

	if err != ErrEmailAlreadyExists {
		t.Errorf("Expected ErrEmailAlreadyExists, got %v", err)
	}
}

// TestGetUserByEmail_Success tests getting user by email
func TestGetUserByEmail_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// Create user
	createdUser, err := svc.RegisterUser(ctx, "test@example.com", "hashedpassword", "Test User")
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}

	// Get user by email
	user, err := svc.GetUserByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, user.ID)
	}
}

// TestGetUserByEmail_NotFound tests getting non-existent user
func TestGetUserByEmail_NotFound(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	_, err := svc.GetUserByEmail(ctx, "nonexistent@example.com")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

// TestIncrementFailedLogin tests incrementing failed login count
func TestIncrementFailedLogin(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// Create user
	user, err := svc.RegisterUser(ctx, "test@example.com", "hashedpassword", "Test User")
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}

	// Increment failed login
	err = svc.IncrementFailedLogin(ctx, user)
	if err != nil {
		t.Fatalf("IncrementFailedLogin failed: %v", err)
	}

	if user.FailedLoginCount != 1 {
		t.Errorf("Expected failed login count 1, got %d", user.FailedLoginCount)
	}
}

// TestIncrementFailedLogin_AccountLock tests account lock after 3 failures
func TestIncrementFailedLogin_AccountLock(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// Create user with 2 failed attempts
	user, err := svc.RegisterUser(ctx, "test@example.com", "hashedpassword", "Test User")
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}
	user.FailedLoginCount = 2

	// Third failed attempt should lock account
	err = svc.IncrementFailedLogin(ctx, user)
	if err != nil {
		t.Fatalf("IncrementFailedLogin failed: %v", err)
	}

	if user.FailedLoginCount != 3 {
		t.Errorf("Expected failed login count 3, got %d", user.FailedLoginCount)
	}

	if user.LockedUntil == nil {
		t.Error("Expected account to be locked")
	}

	// Check lock duration is approximately 30 minutes
	expectedLock := time.Now().Add(30 * time.Minute)
	if user.LockedUntil.Before(expectedLock.Add(-1*time.Minute)) || user.LockedUntil.After(expectedLock.Add(1*time.Minute)) {
		t.Errorf("Expected lock until around %v, got %v", expectedLock, user.LockedUntil)
	}
}

// TestResetFailedLogin tests resetting failed login count
func TestResetFailedLogin(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// Create user with failed attempts
	user, err := svc.RegisterUser(ctx, "test@example.com", "hashedpassword", "Test User")
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}
	user.FailedLoginCount = 2
	lockedUntil := time.Now().Add(30 * time.Minute)
	user.LockedUntil = &lockedUntil

	// Reset failed login
	err = svc.ResetFailedLogin(ctx, user)
	if err != nil {
		t.Fatalf("ResetFailedLogin failed: %v", err)
	}

	if user.FailedLoginCount != 0 {
		t.Errorf("Expected failed login count 0, got %d", user.FailedLoginCount)
	}

	if user.LockedUntil != nil {
		t.Error("Expected LockedUntil to be nil")
	}
}

// TestIsAccountLocked_NotLocked tests account not locked
func TestIsAccountLocked_NotLocked(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)

	user := &model.User{
		Email:       "test@example.com",
		LockedUntil: nil,
	}

	if svc.IsAccountLocked(user) {
		t.Error("Expected account not to be locked")
	}
}

// TestIsAccountLocked_Locked tests account is locked
func TestIsAccountLocked_Locked(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)

	lockedUntil := time.Now().Add(30 * time.Minute)
	user := &model.User{
		Email:       "test@example.com",
		LockedUntil: &lockedUntil,
	}

	if !svc.IsAccountLocked(user) {
		t.Error("Expected account to be locked")
	}
}

// TestIsAccountLocked_Expired tests account lock expired
func TestIsAccountLocked_Expired(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)

	lockedUntil := time.Now().Add(-1 * time.Minute) // Expired
	user := &model.User{
		Email:       "test@example.com",
		LockedUntil: &lockedUntil,
	}

	if svc.IsAccountLocked(user) {
		t.Error("Expected account lock to be expired")
	}
}

// TestGetOrCreateUser_NewUser tests creating new Firebase user
func TestGetOrCreateUser_NewUser(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	user, err := svc.GetOrCreateUser(ctx, "firebase123", "firebase@example.com", "Firebase User", "http://photo.url")
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	if user.FirebaseUID != "firebase123" {
		t.Errorf("Expected FirebaseUID 'firebase123', got '%s'", user.FirebaseUID)
	}

	if user.Email != "firebase@example.com" {
		t.Errorf("Expected email 'firebase@example.com', got '%s'", user.Email)
	}
}

// TestGetOrCreateUser_ExistingUser tests getting existing Firebase user
func TestGetOrCreateUser_ExistingUser(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// Create first user
	firstUser, err := svc.GetOrCreateUser(ctx, "firebase123", "firebase@example.com", "Firebase User", "")
	if err != nil {
		t.Fatalf("First GetOrCreateUser failed: %v", err)
	}

	// Get same user again
	secondUser, err := svc.GetOrCreateUser(ctx, "firebase123", "firebase@example.com", "Updated Name", "")
	if err != nil {
		t.Fatalf("Second GetOrCreateUser failed: %v", err)
	}

	if firstUser.ID != secondUser.ID {
		t.Errorf("Expected same user ID, got %d and %d", firstUser.ID, secondUser.ID)
	}
}

// TestBlacklistToken tests adding token to blacklist
func TestBlacklistToken(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	expiresAt := time.Now().Add(24 * time.Hour)
	err := svc.BlacklistToken(ctx, "tokenhash123", expiresAt)
	if err != nil {
		t.Fatalf("BlacklistToken failed: %v", err)
	}

	// Verify token is blacklisted
	isBlacklisted, err := mockRepos.GetMockTokenBlacklistRepository().IsBlacklisted(ctx, "tokenhash123")
	if err != nil {
		t.Fatalf("IsBlacklisted failed: %v", err)
	}

	if !isBlacklisted {
		t.Error("Expected token to be blacklisted")
	}
}
