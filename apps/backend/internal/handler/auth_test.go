package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/auth"
	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
	"github.com/secure-scorecard/backend/internal/service"
	"github.com/secure-scorecard/backend/internal/validator"
)

// setupTestHandler creates a test handler with mock repositories
func setupTestHandler() (*AuthHandler, *repository.MockRepositories) {
	mockRepos := repository.NewMockRepositories()
	svc := service.NewService(mockRepos)
	jwtManager := auth.NewJWTManager("test-secret-key-for-testing-purposes", 24)
	handler := NewAuthHandler(svc, jwtManager)
	return handler, mockRepos
}

// createTestContext creates an Echo context for testing
func createTestContext(method, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = validator.NewValidator()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// TestRegister_Success tests successful user registration
func TestRegister_Success(t *testing.T) {
	handler, _ := setupTestHandler()

	body := `{"email": "test@example.com", "password": "password123", "display_name": "Test User"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	err := handler.Register(c)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token in response")
	}

	if response.User == nil {
		t.Error("Expected user in response")
	}
}

// TestRegister_DuplicateEmail tests registration with existing email
func TestRegister_DuplicateEmail(t *testing.T) {
	handler, mockRepos := setupTestHandler()

	// Create existing user
	existingUser := &model.User{
		Email:        "existing@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Existing User",
		IsActive:     true,
	}
	mockRepos.GetMockUserRepository().Create(context.Background(), existingUser)

	// Try to register with same email
	body := `{"email": "existing@example.com", "password": "password123", "display_name": "New User"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	err := handler.Register(c)

	// The error should be returned (conflict error)
	if err == nil {
		t.Error("Expected error for duplicate email")
		return
	}

	// Check for 409 Conflict status in error
	if he, ok := err.(*echo.HTTPError); ok {
		if he.Code != http.StatusConflict {
			t.Errorf("Expected status %d, got %d", http.StatusConflict, he.Code)
		}
	} else {
		// Check recorded status
		if rec.Code != http.StatusConflict && rec.Code != http.StatusOK {
			t.Logf("Got error type: %T, value: %v", err, err)
		}
	}
}

// TestRegister_InvalidEmail tests registration with invalid email format
func TestRegister_InvalidEmail(t *testing.T) {
	handler, _ := setupTestHandler()

	body := `{"email": "invalid-email", "password": "password123"}`
	c, _ := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	err := handler.Register(c)

	if err == nil {
		t.Error("Expected error for invalid email")
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		if he.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, he.Code)
		}
	}
}

// TestRegister_WeakPassword tests registration with too short password
func TestRegister_WeakPassword(t *testing.T) {
	handler, _ := setupTestHandler()

	body := `{"email": "test@example.com", "password": "short"}`
	c, _ := createTestContext(http.MethodPost, "/api/v1/auth/register", body)

	err := handler.Register(c)

	if err == nil {
		t.Error("Expected error for weak password")
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		if he.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, he.Code)
		}
	}
}

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	handler, mockRepos := setupTestHandler()

	// Create test user with hashed password
	hashedPassword, _ := auth.HashPassword("password123")
	testUser := &model.User{
		Email:            "test@example.com",
		PasswordHash:     hashedPassword,
		DisplayName:      "Test User",
		IsActive:         true,
		FailedLoginCount: 0,
	}
	mockRepos.GetMockUserRepository().Create(context.Background(), testUser)

	// Login
	body := `{"email": "test@example.com", "password": "password123"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	err := handler.Login(c)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token in response")
	}
}

// TestLogin_InvalidPassword tests login with wrong password
func TestLogin_InvalidPassword(t *testing.T) {
	handler, mockRepos := setupTestHandler()

	// Create test user
	hashedPassword, _ := auth.HashPassword("password123")
	testUser := &model.User{
		Email:            "test@example.com",
		PasswordHash:     hashedPassword,
		DisplayName:      "Test User",
		IsActive:         true,
		FailedLoginCount: 0,
	}
	mockRepos.GetMockUserRepository().Create(context.Background(), testUser)

	// Login with wrong password
	body := `{"email": "test@example.com", "password": "wrongpassword"}`
	c, _ := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	err := handler.Login(c)

	if err == nil {
		t.Error("Expected error for invalid password")
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		if he.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, he.Code)
		}
	}
}

// TestLogin_UserNotFound tests login with non-existent email
func TestLogin_UserNotFound(t *testing.T) {
	handler, _ := setupTestHandler()

	body := `{"email": "nonexistent@example.com", "password": "password123"}`
	c, _ := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	err := handler.Login(c)

	if err == nil {
		t.Error("Expected error for non-existent user")
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		if he.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, he.Code)
		}
	}
}

// TestLogin_AccountLocked tests login with locked account
func TestLogin_AccountLocked(t *testing.T) {
	handler, mockRepos := setupTestHandler()

	// Create test user with locked account
	hashedPassword, _ := auth.HashPassword("password123")
	lockedUntil := time.Now().Add(30 * time.Minute)
	testUser := &model.User{
		Email:            "locked@example.com",
		PasswordHash:     hashedPassword,
		DisplayName:      "Locked User",
		IsActive:         true,
		FailedLoginCount: 3,
		LockedUntil:      &lockedUntil,
	}
	mockRepos.GetMockUserRepository().Create(context.Background(), testUser)

	// Try to login
	body := `{"email": "locked@example.com", "password": "password123"}`
	c, _ := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	err := handler.Login(c)

	if err == nil {
		t.Error("Expected error for locked account")
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		if he.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, he.Code)
		}
	}
}

// TestLogin_FailedLoginIncrement tests that failed login count is incremented
func TestLogin_FailedLoginIncrement(t *testing.T) {
	handler, mockRepos := setupTestHandler()

	// Create test user
	hashedPassword, _ := auth.HashPassword("password123")
	testUser := &model.User{
		Email:            "test@example.com",
		PasswordHash:     hashedPassword,
		DisplayName:      "Test User",
		IsActive:         true,
		FailedLoginCount: 0,
	}
	mockRepos.GetMockUserRepository().Create(context.Background(), testUser)

	// Login with wrong password
	body := `{"email": "test@example.com", "password": "wrongpassword"}`
	c, _ := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	_ = handler.Login(c)

	// Check that failed login count was incremented
	user, _ := mockRepos.GetMockUserRepository().GetByEmail(context.Background(), "test@example.com")
	if user.FailedLoginCount != 1 {
		t.Errorf("Expected failed login count 1, got %d", user.FailedLoginCount)
	}
}

// TestLogin_AccountLockAfterThreeFailures tests account lock after 3 failed attempts
func TestLogin_AccountLockAfterThreeFailures(t *testing.T) {
	handler, mockRepos := setupTestHandler()

	// Create test user
	hashedPassword, _ := auth.HashPassword("password123")
	testUser := &model.User{
		Email:            "test@example.com",
		PasswordHash:     hashedPassword,
		DisplayName:      "Test User",
		IsActive:         true,
		FailedLoginCount: 2, // Already 2 failed attempts
	}
	mockRepos.GetMockUserRepository().Create(context.Background(), testUser)

	// Third failed login attempt
	body := `{"email": "test@example.com", "password": "wrongpassword"}`
	c, _ := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	_ = handler.Login(c)

	// Check that account is now locked
	user, _ := mockRepos.GetMockUserRepository().GetByEmail(context.Background(), "test@example.com")
	if user.FailedLoginCount != 3 {
		t.Errorf("Expected failed login count 3, got %d", user.FailedLoginCount)
	}
	if user.LockedUntil == nil {
		t.Error("Expected account to be locked")
	}
}

// TestLogin_SuccessResetsFailedCount tests that successful login resets failed count
func TestLogin_SuccessResetsFailedCount(t *testing.T) {
	handler, mockRepos := setupTestHandler()

	// Create test user with some failed attempts
	hashedPassword, _ := auth.HashPassword("password123")
	testUser := &model.User{
		Email:            "test@example.com",
		PasswordHash:     hashedPassword,
		DisplayName:      "Test User",
		IsActive:         true,
		FailedLoginCount: 2,
	}
	mockRepos.GetMockUserRepository().Create(context.Background(), testUser)

	// Successful login
	body := `{"email": "test@example.com", "password": "password123"}`
	c, _ := createTestContext(http.MethodPost, "/api/v1/auth/login", body)

	err := handler.Login(c)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Check that failed login count was reset
	user, _ := mockRepos.GetMockUserRepository().GetByEmail(context.Background(), "test@example.com")
	if user.FailedLoginCount != 0 {
		t.Errorf("Expected failed login count 0, got %d", user.FailedLoginCount)
	}
}

// TestFirebaseLogin_NewUser tests Firebase login with new user creation
func TestFirebaseLogin_NewUser(t *testing.T) {
	handler, _ := setupTestHandler()

	body := `{"firebase_uid": "firebase123", "email": "firebase@example.com", "display_name": "Firebase User"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/firebase-login", body)

	err := handler.FirebaseLogin(c)
	if err != nil {
		t.Fatalf("FirebaseLogin failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token in response")
	}
}

// TestFirebaseLogin_ExistingUser tests Firebase login with existing user
func TestFirebaseLogin_ExistingUser(t *testing.T) {
	handler, mockRepos := setupTestHandler()

	// Create existing Firebase user
	existingUser := &model.User{
		FirebaseUID: "firebase123",
		Email:       "firebase@example.com",
		DisplayName: "Firebase User",
		IsActive:    true,
	}
	mockRepos.GetMockUserRepository().Create(context.Background(), existingUser)

	body := `{"firebase_uid": "firebase123", "email": "firebase@example.com", "display_name": "Firebase User"}`
	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/firebase-login", body)

	err := handler.FirebaseLogin(c)
	if err != nil {
		t.Fatalf("FirebaseLogin failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// TestLogout_Success tests successful logout
func TestLogout_Success(t *testing.T) {
	handler, _ := setupTestHandler()

	c, rec := createTestContext(http.MethodPost, "/api/v1/auth/logout", "")

	err := handler.Logout(c)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Logged out successfully" {
		t.Errorf("Expected logout message, got %s", response["message"])
	}
}

// TestLogout_WithToken tests logout with token blacklisting
func TestLogout_WithToken(t *testing.T) {
	handler, mockRepos := setupTestHandler()

	// Create a valid JWT token
	jwtManager := auth.NewJWTManager("test-secret-key-for-testing-purposes", 24)
	token, _ := jwtManager.GenerateToken(1, "firebase123", "test@example.com")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Logout(c)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Check that token was blacklisted
	tokenHash := auth.HashToken(token)
	isBlacklisted, _ := mockRepos.GetMockTokenBlacklistRepository().IsBlacklisted(context.Background(), tokenHash)
	if !isBlacklisted {
		t.Error("Expected token to be blacklisted")
	}
}
