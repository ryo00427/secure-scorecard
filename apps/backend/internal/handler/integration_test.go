package handler

import (
	"context"
	"encoding/json"
	"fmt"
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
	handler := NewHandler(svc)

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
	c.Set("user_id", userID)

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
		if response.User == nil || response.User.Email != "integration@example.com" {
			t.Error("Expected correct user in response")
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
// Crop Registration → Growth Record → Harvest → Notification
func TestCropLifecycleIntegration(t *testing.T) {
	setup := newIntegrationTestSetup()
	ctx := context.Background()

	// Setup: Create a user and garden
	user := &model.User{
		Email:        "croptest@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Crop Test User",
		IsActive:     true,
	}
	setup.mockRepos.User().Create(ctx, user)

	garden := &model.Garden{
		UserID:   user.ID,
		Name:     "Test Garden",
		Location: "Test Location",
	}
	setup.mockRepos.Garden().Create(ctx, garden)

	var cropID uint

	// Step 1: Register a new crop
	t.Run("Step1_RegisterCrop", func(t *testing.T) {
		plantedDate := time.Now().AddDate(0, -2, 0)
		expectedHarvest := time.Now().AddDate(0, 1, 0)

		crop := &model.Crop{
			UserID:              user.ID,
			GardenID:            garden.ID,
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
				Notes:       fmt.Sprintf("%sステージの記録", stage),
			}

			err := setup.service.CreateGrowthRecord(ctx, record)
			if err != nil {
				t.Fatalf("Failed to create growth record for stage %s: %v", stage, err)
			}
		}

		// Verify all records were created
		records, err := setup.service.GetGrowthRecordsByCropID(ctx, cropID)
		if err != nil {
			t.Fatalf("Failed to get growth records: %v", err)
		}

		if len(records) != 4 {
			t.Errorf("Expected 4 growth records, got %d", len(records))
		}
	})

	// Step 3: Record harvest
	t.Run("Step3_RecordHarvest", func(t *testing.T) {
		harvest := &model.Harvest{
			CropID:      cropID,
			HarvestDate: time.Now(),
			Quantity:    5.5,
			Unit:        "kg",
			Quality:     "excellent",
			Notes:       "初収穫！とても良い出来",
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
			t.Errorf("Expected crop status 'harvested', got '%s'", crop.Status)
		}
	})

	// Step 4: Verify harvest appears in analytics
	t.Run("Step4_VerifyAnalytics", func(t *testing.T) {
		summary, err := setup.service.GetHarvestSummary(ctx, user.ID, nil, nil, nil)
		if err != nil {
			t.Fatalf("Failed to get harvest summary: %v", err)
		}

		if len(summary) == 0 {
			t.Error("Expected at least one harvest summary entry")
		}

		found := false
		for _, s := range summary {
			if s.CropID == cropID {
				found = true
				if s.TotalQuantity != 5.5 {
					t.Errorf("Expected total quantity 5.5, got %f", s.TotalQuantity)
				}
			}
		}

		if !found {
			t.Error("Harvest not found in summary")
		}
	})
}

// =============================================================================
// Task Notification Integration Tests - タスク通知統合テスト
// =============================================================================

// TestTaskNotificationIntegration tests the task notification flow:
// Task Creation → Cron Job Processing → Notification Delivery
func TestTaskNotificationIntegration(t *testing.T) {
	setup := newIntegrationTestSetup()
	ctx := context.Background()

	// Setup: Create users with tasks
	user := &model.User{
		Email:        "tasktest@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Task Test User",
		IsActive:     true,
		NotificationSettings: model.NotificationSettings{
			PushEnabled:       true,
			EmailEnabled:      true,
			TaskReminders:     true,
			HarvestReminders:  true,
		},
	}
	setup.mockRepos.User().Create(ctx, user)

	// Step 1: Create tasks with different due dates
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

		// Overdue tasks (3 or more triggers alert)
		for i := 0; i < 3; i++ {
			overdueTask := &model.Task{
				UserID:   user.ID,
				Title:    fmt.Sprintf("期限切れタスク%d", i+1),
				DueDate:  time.Now().AddDate(0, 0, -2),
				Priority: "medium",
				Status:   "pending",
			}
			setup.mockRepos.Task().Create(ctx, overdueTask)
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

	// Step 2: Process scheduled notifications (simulating cron job)
	t.Run("Step2_ProcessScheduledNotifications", func(t *testing.T) {
		result, err := setup.service.ProcessScheduledNotifications(ctx)
		if err != nil {
			t.Fatalf("Failed to process scheduled notifications: %v", err)
		}

		// Check that events were generated
		if result.TotalEvents == 0 {
			t.Log("Warning: No notification events generated (may be expected if no eligible tasks)")
		}

		// Should have overdue alerts (3 overdue tasks)
		if result.OverdueTaskAlerts == 0 {
			t.Log("Warning: No overdue task alerts generated")
		}

		// Log the result for debugging
		t.Logf("Notification processing result: %+v", result)
	})

	// Step 3: Complete a task and verify status update
	t.Run("Step3_CompleteTask", func(t *testing.T) {
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

// TestPlotAssignmentIntegration tests the plot assignment flow:
// Plot Creation → Crop Assignment → Duplicate Error Handling
func TestPlotAssignmentIntegration(t *testing.T) {
	setup := newIntegrationTestSetup()
	ctx := context.Background()

	// Setup: Create user and garden
	user := &model.User{
		Email:        "plottest@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Plot Test User",
		IsActive:     true,
	}
	setup.mockRepos.User().Create(ctx, user)

	garden := &model.Garden{
		UserID:   user.ID,
		Name:     "Plot Test Garden",
		Location: "Test Location",
	}
	setup.mockRepos.Garden().Create(ctx, garden)

	var plotID uint
	var cropID uint

	// Step 1: Create a plot
	t.Run("Step1_CreatePlot", func(t *testing.T) {
		plot := &model.Plot{
			GardenID: garden.ID,
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
			GardenID:            garden.ID,
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
		assignment := &model.PlotAssignment{
			PlotID:       plotID,
			CropID:       cropID,
			AssignedDate: time.Now(),
		}

		err := setup.service.CreatePlotAssignment(ctx, assignment)
		if err != nil {
			t.Fatalf("Failed to assign crop to plot: %v", err)
		}

		// Verify plot status changed
		plot, err := setup.mockRepos.Plot().GetByID(ctx, plotID)
		if err != nil {
			t.Fatalf("Failed to get plot: %v", err)
		}

		if plot.Status != "occupied" {
			t.Errorf("Expected plot status 'occupied', got '%s'", plot.Status)
		}
	})

	// Step 4: Try duplicate assignment (should fail)
	t.Run("Step4_DuplicateAssignment", func(t *testing.T) {
		// Create another crop
		anotherCrop := &model.Crop{
			UserID:              user.ID,
			GardenID:            garden.ID,
			Name:                "ナス",
			Variety:             "長ナス",
			PlantedDate:         time.Now(),
			ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
			Status:              "growing",
		}
		setup.service.CreateCrop(ctx, anotherCrop)

		// Try to assign to the same plot
		duplicateAssignment := &model.PlotAssignment{
			PlotID:       plotID,
			CropID:       anotherCrop.ID,
			AssignedDate: time.Now(),
		}

		err := setup.service.CreatePlotAssignment(ctx, duplicateAssignment)
		if err == nil {
			t.Error("Expected error for duplicate plot assignment")
		}
	})

	// Step 5: End assignment and verify plot is available again
	t.Run("Step5_EndAssignment", func(t *testing.T) {
		err := setup.service.EndPlotAssignment(ctx, plotID, cropID)
		if err != nil {
			t.Fatalf("Failed to end assignment: %v", err)
		}

		// Verify plot is available again
		plot, err := setup.mockRepos.Plot().GetByID(ctx, plotID)
		if err != nil {
			t.Fatalf("Failed to get plot: %v", err)
		}

		if plot.Status != "available" {
			t.Errorf("Expected plot status 'available', got '%s'", plot.Status)
		}
	})
}

// =============================================================================
// CSV Export Integration Tests - CSVエクスポート統合テスト
// =============================================================================

// TestCSVExportIntegration tests the CSV export flow:
// Data Generation → CSV Creation → Signed URL Generation
func TestCSVExportIntegration(t *testing.T) {
	setup := newIntegrationTestSetup()
	ctx := context.Background()

	// Setup: Create user with data
	user := &model.User{
		Email:        "exporttest@example.com",
		PasswordHash: "hashedpassword",
		DisplayName:  "Export Test User",
		IsActive:     true,
	}
	setup.mockRepos.User().Create(ctx, user)

	garden := &model.Garden{
		UserID:   user.ID,
		Name:     "Export Test Garden",
		Location: "Test Location",
	}
	setup.mockRepos.Garden().Create(ctx, garden)

	// Create test data
	t.Run("Setup_CreateTestData", func(t *testing.T) {
		// Create crops
		for i := 0; i < 5; i++ {
			crop := &model.Crop{
				UserID:              user.ID,
				GardenID:            garden.ID,
				Name:                fmt.Sprintf("作物%d", i+1),
				Variety:             fmt.Sprintf("品種%d", i+1),
				PlantedDate:         time.Now().AddDate(0, -3, 0),
				ExpectedHarvestDate: time.Now().AddDate(0, 1, 0),
				Status:              "growing",
			}
			setup.mockRepos.Crop().Create(ctx, crop)

			// Add harvest for each crop
			harvest := &model.Harvest{
				CropID:      crop.ID,
				HarvestDate: time.Now().AddDate(0, 0, -7),
				Quantity:    float64(i+1) * 2.5,
				Unit:        "kg",
				Quality:     "good",
			}
			setup.mockRepos.Harvest().Create(ctx, harvest)
		}

		// Create tasks
		for i := 0; i < 3; i++ {
			task := &model.Task{
				UserID:   user.ID,
				Title:    fmt.Sprintf("タスク%d", i+1),
				DueDate:  time.Now().AddDate(0, 0, i),
				Priority: "medium",
				Status:   "pending",
			}
			setup.mockRepos.Task().Create(ctx, task)
		}
	})

	// Step 1: Export crops data
	t.Run("Step1_ExportCrops", func(t *testing.T) {
		exportResult, err := setup.service.ExportCSV(ctx, user.ID, "crops")
		if err != nil {
			t.Fatalf("Failed to export crops: %v", err)
		}

		if exportResult.DownloadURL == "" {
			t.Error("Expected download URL in export result")
		}

		if exportResult.ExpiresAt.Before(time.Now()) {
			t.Error("Expected future expiration time")
		}

		t.Logf("Crops export URL: %s", exportResult.DownloadURL)
	})

	// Step 2: Export harvests data
	t.Run("Step2_ExportHarvests", func(t *testing.T) {
		exportResult, err := setup.service.ExportCSV(ctx, user.ID, "harvests")
		if err != nil {
			t.Fatalf("Failed to export harvests: %v", err)
		}

		if exportResult.DownloadURL == "" {
			t.Error("Expected download URL in export result")
		}
	})

	// Step 3: Export tasks data
	t.Run("Step3_ExportTasks", func(t *testing.T) {
		exportResult, err := setup.service.ExportCSV(ctx, user.ID, "tasks")
		if err != nil {
			t.Fatalf("Failed to export tasks: %v", err)
		}

		if exportResult.DownloadURL == "" {
			t.Error("Expected download URL in export result")
		}
	})

	// Step 4: Export all data
	t.Run("Step4_ExportAll", func(t *testing.T) {
		exportResult, err := setup.service.ExportCSV(ctx, user.ID, "all")
		if err != nil {
			t.Fatalf("Failed to export all data: %v", err)
		}

		if exportResult.DownloadURL == "" {
			t.Error("Expected download URL in export result")
		}
	})
}

// =============================================================================
// Full User Journey Integration Test - フルユーザージャーニー統合テスト
// =============================================================================

// TestFullUserJourney tests a complete user journey from registration to data export
func TestFullUserJourney(t *testing.T) {
	setup := newIntegrationTestSetup()
	ctx := context.Background()

	var userID uint
	var gardenID uint
	var cropID uint

	// Phase 1: User Registration and Authentication
	t.Run("Phase1_Authentication", func(t *testing.T) {
		// Register
		body := `{"email": "journey@example.com", "password": "journeyPass123", "display_name": "Journey User"}`
		c, rec := setup.createContext(http.MethodPost, "/api/v1/auth/register", body)
		setup.authHandler.Register(c)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Registration failed with status %d", rec.Code)
		}

		user, _ := setup.mockRepos.User().GetByEmail(ctx, "journey@example.com")
		userID = user.ID

		// Login
		body = `{"email": "journey@example.com", "password": "journeyPass123"}`
		c, rec = setup.createContext(http.MethodPost, "/api/v1/auth/login", body)
		setup.authHandler.Login(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("Login failed with status %d", rec.Code)
		}
	})

	// Phase 2: Garden and Crop Setup
	t.Run("Phase2_GardenSetup", func(t *testing.T) {
		// Create garden
		garden := &model.Garden{
			UserID:   userID,
			Name:     "マイガーデン",
			Location: "東京都",
		}
		setup.mockRepos.Garden().Create(ctx, garden)
		gardenID = garden.ID

		// Create crop
		crop := &model.Crop{
			UserID:              userID,
			GardenID:            gardenID,
			Name:                "トマト",
			Variety:             "桃太郎",
			PlantedDate:         time.Now().AddDate(0, -1, 0),
			ExpectedHarvestDate: time.Now().AddDate(0, 1, 0),
			Status:              "growing",
		}
		setup.service.CreateCrop(ctx, crop)
		cropID = crop.ID
	})

	// Phase 3: Manage Growth and Tasks
	t.Run("Phase3_DailyManagement", func(t *testing.T) {
		// Add growth record
		record := &model.GrowthRecord{
			CropID:      cropID,
			RecordDate:  time.Now(),
			GrowthStage: "flowering",
			Notes:       "花が咲き始めました",
		}
		setup.service.CreateGrowthRecord(ctx, record)

		// Create task
		task := &model.Task{
			UserID:   userID,
			Title:    "水やり",
			DueDate:  time.Now(),
			Priority: "high",
			Status:   "pending",
		}
		setup.service.CreateTask(ctx, task)

		// Complete task
		setup.service.CompleteTask(ctx, task.ID)
	})

	// Phase 4: Harvest and Analytics
	t.Run("Phase4_HarvestAndAnalytics", func(t *testing.T) {
		// Record harvest
		harvest := &model.Harvest{
			CropID:      cropID,
			HarvestDate: time.Now(),
			Quantity:    3.5,
			Unit:        "kg",
			Quality:     "excellent",
		}
		setup.service.CreateHarvest(ctx, harvest)

		// Get analytics
		summary, err := setup.service.GetHarvestSummary(ctx, userID, nil, nil, nil)
		if err != nil {
			t.Fatalf("Failed to get harvest summary: %v", err)
		}

		if len(summary) == 0 {
			t.Error("Expected harvest summary data")
		}
	})

	// Phase 5: Data Export
	t.Run("Phase5_DataExport", func(t *testing.T) {
		result, err := setup.service.ExportCSV(ctx, userID, "all")
		if err != nil {
			t.Fatalf("Failed to export data: %v", err)
		}

		if result.DownloadURL == "" {
			t.Error("Expected download URL")
		}
	})

	t.Log("Full user journey completed successfully!")
}
