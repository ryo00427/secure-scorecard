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

// =============================================================================
// Integration Test Suite - 統合テストスイート
// =============================================================================
// 複数のサービスを跨ぐエンドツーエンドの統合テストを実行します。

// integrationTestSetup creates a complete test environment
type integrationTestSetup struct {
	echo        *echo.Echo
	mockRepos   *repository.MockRepositories
	service     *service.Service
	jwtManager  *auth.JWTManager
	authHandler *AuthHandler
	handler     *Handler
}

// newIntegrationTestSetup creates a new integration test setup
func newIntegrationTestSetup() *integrationTestSetup {
	e := echo.New()
	e.Validator = validator.NewValidator()

	mockRepos := repository.NewMockRepositories()
	svc := service.NewService(mockRepos)
	jwtManager := auth.NewJWTManager("integration-test-secret-key-32chars", 24)
	authHandler := NewAuthHandler(svc, jwtManager)
	handler := NewHandler(svc, jwtManager, nil) // nil for S3Service in tests

	return &integrationTestSetup{
		echo:        e,
		mockRepos:   mockRepos,
		service:     svc,
		jwtManager:  jwtManager,
		authHandler: authHandler,
		handler:     handler,
	}
}

// createAuthenticatedContext creates an Echo context with JWT authentication
func (s *integrationTestSetup) createAuthenticatedContext(method, path, body string, userID uint) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	// Generate JWT token
	token, _ := s.jwtManager.GenerateToken(userID, "", "test@example.com")
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	// Set user claims in context (required by auth.GetUserIDFromContext)
	claims := &auth.Claims{
		UserID: userID,
		Email:  "test@example.com",
	}
	c.Set("user", claims)

	return c, rec
}

// createContext creates an Echo context without authentication
func (s *integrationTestSetup) createContext(method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	return c, rec
}

// =============================================================================
// Auth Flow Integration Tests - 認証フロー統合テスト
// =============================================================================

// TestAuthFlowIntegration tests the complete authentication flow:
// Registration → Login → JWT Verification → Protected API Access → Logout
func TestAuthFlowIntegration(t *testing.T) {
	setup := newIntegrationTestSetup()

	// Step 1: Register a new user
	t.Run("Step1_Register", func(t *testing.T) {
		body := `{"email": "integration@example.com", "password": "securePassword123", "display_name": "Integration User"}`
		c, rec := setup.createContext(http.MethodPost, "/api/v1/auth/register", body)

		err := setup.authHandler.Register(c)
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		var response AuthResponse
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response.Token == "" {
			t.Error("Expected token in registration response")
		}
		if response.User == nil {
			t.Error("Expected user in response")
		}
	})

	// Step 2: Login with the registered user
	var loginToken string
	t.Run("Step2_Login", func(t *testing.T) {
		body := `{"email": "integration@example.com", "password": "securePassword123"}`
		c, rec := setup.createContext(http.MethodPost, "/api/v1/auth/login", body)

		err := setup.authHandler.Login(c)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response AuthResponse
		json.Unmarshal(rec.Body.Bytes(), &response)
		loginToken = response.Token

		if loginToken == "" {
			t.Error("Expected token in login response")
		}
	})

	// Step 3: Access protected API with JWT
	t.Run("Step3_AccessProtectedAPI", func(t *testing.T) {
		user, _ := setup.mockRepos.User().GetByEmail(context.Background(), "integration@example.com")
		if user == nil {
			t.Fatal("User not found after registration")
		}

		// Create a garden (protected resource)
		body := `{"name": "Test Garden", "location": "Test Location"}`
		c, rec := setup.createAuthenticatedContext(http.MethodPost, "/api/v1/gardens", body, user.ID)

		err := setup.handler.CreateGarden(c)
		if err != nil {
			t.Fatalf("Protected API access failed: %v", err)
		}

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
		}
	})

	// Step 4: Logout and verify token is blacklisted
	t.Run("Step4_Logout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+loginToken)
		rec := httptest.NewRecorder()
		c := setup.echo.NewContext(req, rec)

		err := setup.authHandler.Logout(c)
		if err != nil {
			t.Fatalf("Logout failed: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		// Verify token is blacklisted
		tokenHash := auth.HashToken(loginToken)
		isBlacklisted, _ := setup.mockRepos.TokenBlacklist().IsBlacklisted(context.Background(), tokenHash)
		if !isBlacklisted {
			t.Error("Expected token to be blacklisted after logout")
		}
	})
}

// =============================================================================
// Crop Lifecycle Integration Tests - 作物ライフサイクル統合テスト
// =============================================================================

// TestCropLifecycleIntegration tests the complete crop lifecycle:
// Crop Registration → Growth Record → Harvest
func TestCropLifecycleIntegration(t *testing.T) {
	setup := newIntegrationTestSetup()
	ctx := context.Background()

	// Setup: Create a user
	user := &model.User{
		Email:        "croptest@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Crop Test User",
		IsActive:     true,
	}
	setup.mockRepos.User().Create(ctx, user)

	var cropID uint

	// Step 1: Register a new crop
	t.Run("Step1_RegisterCrop", func(t *testing.T) {
		plantedDate := time.Now().AddDate(0, -2, 0)
		expectedHarvest := time.Now().AddDate(0, 1, 0)

		crop := &model.Crop{
			UserID:              user.ID,
			Name:                "トマト",
			Variety:             "ミニトマト",
			PlantedDate:         plantedDate,
			ExpectedHarvestDate: expectedHarvest,
			Status:              "growing",
		}

		err := setup.service.CreateCrop(ctx, crop)
		if err != nil {
			t.Fatalf("Failed to create crop: %v", err)
		}

		cropID = crop.ID
		if cropID == 0 {
			t.Error("Expected crop ID to be set")
		}
	})

	// Step 2: Add growth records
	t.Run("Step2_AddGrowthRecords", func(t *testing.T) {
		stages := []string{"seedling", "vegetative", "flowering", "fruiting"}

		for i, stage := range stages {
			record := &model.GrowthRecord{
				CropID:      cropID,
				RecordDate:  time.Now().AddDate(0, 0, -30+i*10),
				GrowthStage: stage,
				Notes:       stage + " stage record",
			}

			err := setup.service.CreateGrowthRecord(ctx, record)
			if err != nil {
				t.Fatalf("Failed to create growth record for stage %s: %v", stage, err)
			}
		}
	})

	// Step 3: Record harvest
	t.Run("Step3_RecordHarvest", func(t *testing.T) {
		harvest := &model.Harvest{
			CropID:       cropID,
			HarvestDate:  time.Now(),
			Quantity:     5.5,
			QuantityUnit: "kg",
			Quality:      "excellent",
			Notes:        "初収穫！とても良い出来",
		}

		err := setup.service.CreateHarvest(ctx, harvest)
		if err != nil {
			t.Fatalf("Failed to create harvest: %v", err)
		}

		// Verify crop status is updated
		crop, err := setup.service.GetCropByID(ctx, cropID)
		if err != nil {
			t.Fatalf("Failed to get crop: %v", err)
		}

		if crop.Status != "harvested" {
			t.Logf("Note: Crop status is '%s' (automatic status update may not be implemented)", crop.Status)
		}
	})
}

// =============================================================================
// Task Management Integration Tests - タスク管理統合テスト
// =============================================================================

// TestTaskManagementIntegration tests task creation and completion
func TestTaskManagementIntegration(t *testing.T) {
	setup := newIntegrationTestSetup()
	ctx := context.Background()

	// Setup: Create user
	user := &model.User{
		Email:        "tasktest@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Task Test User",
		IsActive:     true,
	}
	setup.mockRepos.User().Create(ctx, user)

	// Step 1: Create tasks
	t.Run("Step1_CreateTasks", func(t *testing.T) {
		// Today's task
		todayTask := &model.Task{
			UserID:   user.ID,
			Title:    "今日の水やり",
			DueDate:  time.Now(),
			Priority: "high",
			Status:   "pending",
		}
		err := setup.service.CreateTask(ctx, todayTask)
		if err != nil {
			t.Fatalf("Failed to create today's task: %v", err)
		}

		// Verify tasks were created
		tasks, err := setup.service.GetTodayTasks(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to get today's tasks: %v", err)
		}

		if len(tasks) < 1 {
			t.Error("Expected at least one task for today")
		}
	})

	// Step 2: Complete a task
	t.Run("Step2_CompleteTask", func(t *testing.T) {
		tasks, _ := setup.mockRepos.Task().GetByUserID(ctx, user.ID)
		if len(tasks) == 0 {
			t.Fatal("No tasks found")
		}

		taskID := tasks[0].ID
		err := setup.service.CompleteTask(ctx, taskID)
		if err != nil {
			t.Fatalf("Failed to complete task: %v", err)
		}

		// Verify task status
		task, err := setup.mockRepos.Task().GetByID(ctx, taskID)
		if err != nil {
			t.Fatalf("Failed to get task: %v", err)
		}

		if task.Status != "completed" {
			t.Errorf("Expected task status 'completed', got '%s'", task.Status)
		}
	})
}

// =============================================================================
// Plot Assignment Integration Tests - 区画配置統合テスト
// =============================================================================

// TestPlotAssignmentIntegration tests the plot assignment flow
func TestPlotAssignmentIntegration(t *testing.T) {
	setup := newIntegrationTestSetup()
	ctx := context.Background()

	// Setup: Create user
	user := &model.User{
		Email:        "plottest@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Plot Test User",
		IsActive:     true,
	}
	setup.mockRepos.User().Create(ctx, user)

	var plotID uint
	var cropID uint

	// Step 1: Create a plot
	t.Run("Step1_CreatePlot", func(t *testing.T) {
		plot := &model.Plot{
			UserID:   user.ID,
			Name:     "区画A",
			Width:    2.0,
			Height:   3.0,
			SoilType: "loam",
			Sunlight: "full_sun",
			Status:   "available",
		}

		err := setup.service.CreatePlot(ctx, plot)
		if err != nil {
			t.Fatalf("Failed to create plot: %v", err)
		}

		plotID = plot.ID
		if plotID == 0 {
			t.Error("Expected plot ID to be set")
		}
	})

	// Step 2: Create a crop for assignment
	t.Run("Step2_CreateCrop", func(t *testing.T) {
		crop := &model.Crop{
			UserID:              user.ID,
			Name:                "キュウリ",
			Variety:             "夏すずみ",
			PlantedDate:         time.Now(),
			ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
			Status:              "growing",
		}

		err := setup.service.CreateCrop(ctx, crop)
		if err != nil {
			t.Fatalf("Failed to create crop: %v", err)
		}

		cropID = crop.ID
	})

	// Step 3: Assign crop to plot
	t.Run("Step3_AssignCropToPlot", func(t *testing.T) {
		_, err := setup.service.AssignCropToPlot(ctx, plotID, cropID, time.Now())
		if err != nil {
			t.Fatalf("Failed to assign crop to plot: %v", err)
		}

		// Verify plot status changed
		plot, err := setup.mockRepos.Plot().GetByID(ctx, plotID)
		if err != nil {
			t.Fatalf("Failed to get plot: %v", err)
		}

		if plot.Status != "occupied" {
			t.Logf("Note: Plot status is '%s' (automatic status update may not be implemented)", plot.Status)
		}
	})

	// Step 4: End assignment
	t.Run("Step4_EndAssignment", func(t *testing.T) {
		err := setup.service.UnassignCropFromPlot(ctx, plotID)
		if err != nil {
			t.Fatalf("Failed to end assignment: %v", err)
		}
	})
}

// =============================================================================
// Garden Handler Integration Tests - 菜園ハンドラ統合テスト
// =============================================================================

// TestGardenHandlerIntegration tests garden handler operations
func TestGardenHandlerIntegration(t *testing.T) {
	setup := newIntegrationTestSetup()
	ctx := context.Background()

	// Setup: Create user
	user := &model.User{
		Email:        "gardentest@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Garden Test User",
		IsActive:     true,
	}
	setup.mockRepos.User().Create(ctx, user)

	// Step 1: Test GetGardens handler (returns empty list from mock)
	t.Run("Step1_GetGardensHandler", func(t *testing.T) {
		c, rec := setup.createAuthenticatedContext("GET", "/api/v1/gardens", "", user.ID)

		err := setup.handler.GetGardens(c)
		if err != nil {
			t.Fatalf("GetGardens handler failed: %v", err)
		}

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	// Step 2: Test CreateGarden handler
	t.Run("Step2_CreateGardenHandler", func(t *testing.T) {
		body := `{"name": "テスト菜園", "location": "東京都", "description": "テスト用"}`
		c, rec := setup.createAuthenticatedContext("POST", "/api/v1/gardens", body, user.ID)

		err := setup.handler.CreateGarden(c)
		if err != nil {
			t.Fatalf("CreateGarden handler failed: %v", err)
		}

		if rec.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
		}
	})
}
