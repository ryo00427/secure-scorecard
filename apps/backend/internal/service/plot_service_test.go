// Package service - PlotService Unit Tests
//
// PlotServiceのユニットテストを提供します。
// MockRepositoryを使用して、データベースなしでサービス層のロジックをテストします。
//
// テスト対象:
//   - 区画CRUD操作
//   - 作物配置・解除操作
//   - レイアウト・履歴取得
//   - データ分離（ユーザーIDによるフィルタリング）
package service

import (
	"context"
	"testing"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
)

// =============================================================================
// CreatePlot テスト
// =============================================================================

// TestCreatePlot_Success は区画の正常作成をテストします。
func TestCreatePlot_Success(t *testing.T) {
	// Arrange: モックリポジトリとサービスをセットアップ
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// テスト用の区画データ
	plot := &model.Plot{
		UserID:   1,
		Name:     "畑A",
		Width:    2.0,
		Height:   3.0,
		SoilType: "loamy",
		Sunlight: "full_sun",
		Status:   "available",
		Notes:    "日当たり良好",
	}

	// Act: 区画を作成
	err := svc.CreatePlot(ctx, plot)

	// Assert: エラーがないことを確認
	if err != nil {
		t.Fatalf("CreatePlot failed: %v", err)
	}

	// IDが割り当てられていることを確認
	if plot.ID == 0 {
		t.Error("Expected plot ID to be assigned, got 0")
	}

	// モックに保存されていることを確認
	savedPlot, err := mockRepos.GetMockPlotRepository().GetByID(ctx, plot.ID)
	if err != nil {
		t.Fatalf("Failed to get saved plot: %v", err)
	}

	if savedPlot.Name != "畑A" {
		t.Errorf("Expected name '畑A', got '%s'", savedPlot.Name)
	}
}

// TestCreatePlot_WithPosition は位置情報付きの区画作成をテストします。
func TestCreatePlot_WithPosition(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	posX := 0
	posY := 1
	plot := &model.Plot{
		UserID:    1,
		Name:      "畑B",
		Width:     1.5,
		Height:    2.0,
		PositionX: &posX,
		PositionY: &posY,
		Status:    "available",
	}

	err := svc.CreatePlot(ctx, plot)

	if err != nil {
		t.Fatalf("CreatePlot failed: %v", err)
	}

	if plot.PositionX == nil || *plot.PositionX != 0 {
		t.Error("Expected PositionX to be 0")
	}
	if plot.PositionY == nil || *plot.PositionY != 1 {
		t.Error("Expected PositionY to be 1")
	}
}

// TestCreatePlot_AllSoilTypes は全ての土壌タイプをテストします。
func TestCreatePlot_AllSoilTypes(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 全ての有効な土壌タイプ
	validSoilTypes := []string{"clay", "sandy", "loamy", "peaty"}

	for _, soilType := range validSoilTypes {
		plot := &model.Plot{
			UserID:   1,
			Name:     "区画_" + soilType,
			Width:    1.0,
			Height:   1.0,
			SoilType: soilType,
			Status:   "available",
		}
		err := svc.CreatePlot(ctx, plot)
		if err != nil {
			t.Errorf("Failed to create plot with soil type '%s': %v", soilType, err)
		}
	}
}

// TestCreatePlot_AllSunlightConditions は全ての日当たり条件をテストします。
func TestCreatePlot_AllSunlightConditions(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 全ての有効な日当たり条件
	validSunlightConditions := []string{"full_sun", "partial_shade", "shade"}

	for _, sunlight := range validSunlightConditions {
		plot := &model.Plot{
			UserID:   1,
			Name:     "区画_" + sunlight,
			Width:    1.0,
			Height:   1.0,
			Sunlight: sunlight,
			Status:   "available",
		}
		err := svc.CreatePlot(ctx, plot)
		if err != nil {
			t.Errorf("Failed to create plot with sunlight '%s': %v", sunlight, err)
		}
	}
}

// =============================================================================
// GetUserPlots テスト
// =============================================================================

// TestGetUserPlots_Success はユーザーの区画一覧取得をテストします。
func TestGetUserPlots_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// テストデータを準備
	userID := uint(1)
	plots := []*model.Plot{
		{
			UserID:   userID,
			Name:     "畑A",
			Width:    2.0,
			Height:   3.0,
			Status:   "available",
		},
		{
			UserID:   userID,
			Name:     "畑B",
			Width:    1.5,
			Height:   2.5,
			Status:   "occupied",
		},
	}

	// モックに区画を追加
	for _, plot := range plots {
		_ = svc.CreatePlot(ctx, plot)
	}

	// Act: ユーザーの区画を取得
	result, err := svc.GetUserPlots(ctx, userID)

	// Assert
	if err != nil {
		t.Fatalf("GetUserPlots failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 plots, got %d", len(result))
	}
}

// TestGetUserPlots_Empty はユーザーに区画がない場合をテストします。
func TestGetUserPlots_Empty(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	result, err := svc.GetUserPlots(ctx, 999) // 存在しないユーザー

	if err != nil {
		t.Fatalf("GetUserPlots failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 plots, got %d", len(result))
	}
}

// =============================================================================
// GetUserPlotsByStatus テスト
// =============================================================================

// TestGetUserPlotsByStatus_Success はステータスでフィルタリングした区画取得をテストします。
func TestGetUserPlotsByStatus_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 異なるステータスの区画を作成
	_ = svc.CreatePlot(ctx, &model.Plot{
		UserID: userID,
		Name:   "畑A",
		Width:  1.0,
		Height: 1.0,
		Status: "available",
	})
	_ = svc.CreatePlot(ctx, &model.Plot{
		UserID: userID,
		Name:   "畑B",
		Width:  1.0,
		Height: 1.0,
		Status: "occupied",
	})
	_ = svc.CreatePlot(ctx, &model.Plot{
		UserID: userID,
		Name:   "畑C",
		Width:  1.0,
		Height: 1.0,
		Status: "occupied",
	})

	// Act: "occupied" ステータスでフィルタ
	result, err := svc.GetUserPlotsByStatus(ctx, userID, "occupied")

	// Assert
	if err != nil {
		t.Fatalf("GetUserPlotsByStatus failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 occupied plots, got %d", len(result))
	}

	for _, plot := range result {
		if plot.Status != "occupied" {
			t.Errorf("Expected status 'occupied', got '%s'", plot.Status)
		}
	}
}

// TestGetUserPlotsByStatus_Available は空き区画の取得をテストします。
func TestGetUserPlotsByStatus_Available(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	_ = svc.CreatePlot(ctx, &model.Plot{
		UserID: userID,
		Name:   "畑A",
		Width:  1.0,
		Height: 1.0,
		Status: "available",
	})
	_ = svc.CreatePlot(ctx, &model.Plot{
		UserID: userID,
		Name:   "畑B",
		Width:  1.0,
		Height: 1.0,
		Status: "occupied",
	})

	result, err := svc.GetUserPlotsByStatus(ctx, userID, "available")

	if err != nil {
		t.Fatalf("GetUserPlotsByStatus failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 available plot, got %d", len(result))
	}
}

// =============================================================================
// UpdatePlot テスト
// =============================================================================

// TestUpdatePlot_Success は区画の更新をテストします。
func TestUpdatePlot_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 区画を作成
	plot := &model.Plot{
		UserID:   1,
		Name:     "畑A",
		Width:    2.0,
		Height:   3.0,
		SoilType: "loamy",
		Status:   "available",
	}
	_ = svc.CreatePlot(ctx, plot)

	// Act: 名前と土壌タイプを更新
	plot.Name = "畑A（改良済み）"
	plot.SoilType = "sandy"
	err := svc.UpdatePlot(ctx, plot)

	// Assert
	if err != nil {
		t.Fatalf("UpdatePlot failed: %v", err)
	}

	// 更新が反映されていることを確認
	updated, _ := svc.GetPlotByID(ctx, plot.ID)
	if updated.Name != "畑A（改良済み）" {
		t.Errorf("Expected name '畑A（改良済み）', got '%s'", updated.Name)
	}
	if updated.SoilType != "sandy" {
		t.Errorf("Expected soil type 'sandy', got '%s'", updated.SoilType)
	}
}

// =============================================================================
// DeletePlot テスト
// =============================================================================

// TestDeletePlot_Success は区画の削除をテストします。
func TestDeletePlot_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 区画を作成
	plot := &model.Plot{
		UserID: 1,
		Name:   "畑A",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	}
	_ = svc.CreatePlot(ctx, plot)

	// Act: 削除
	err := svc.DeletePlot(ctx, plot.ID)

	// Assert
	if err != nil {
		t.Fatalf("DeletePlot failed: %v", err)
	}

	// 削除されていることを確認
	_, err = svc.GetPlotByID(ctx, plot.ID)
	if err == nil {
		t.Error("Expected plot to be deleted, but it still exists")
	}
}

// TestDeletePlot_WithAssignments は配置履歴も含めて削除されることをテストします。
func TestDeletePlot_WithAssignments(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 区画を作成
	plot := &model.Plot{
		UserID: 1,
		Name:   "畑A",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	}
	_ = svc.CreatePlot(ctx, plot)

	// 作物を作成
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 区画に作物を配置
	_, _ = svc.AssignCropToPlot(ctx, plot.ID, crop.ID, time.Now())

	// 配置があることを確認
	assignments, _ := svc.GetPlotAssignments(ctx, plot.ID)
	if len(assignments) != 1 {
		t.Fatalf("Expected 1 assignment, got %d", len(assignments))
	}

	// Act: 区画を削除（関連する配置履歴も削除される）
	err := svc.DeletePlot(ctx, plot.ID)

	// Assert
	if err != nil {
		t.Fatalf("DeletePlot failed: %v", err)
	}

	// 配置履歴も削除されていることを確認
	assignmentsAfter, _ := svc.GetPlotAssignments(ctx, plot.ID)
	if len(assignmentsAfter) != 0 {
		t.Errorf("Expected 0 assignments after deletion, got %d", len(assignmentsAfter))
	}
}

// =============================================================================
// AssignCropToPlot テスト
// =============================================================================

// TestAssignCropToPlot_Success は作物配置の正常動作をテストします。
func TestAssignCropToPlot_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 区画を作成
	plot := &model.Plot{
		UserID: 1,
		Name:   "畑A",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	}
	_ = svc.CreatePlot(ctx, plot)

	// 作物を作成
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	}
	_ = svc.CreateCrop(ctx, crop)

	// Act: 作物を配置
	assignment, err := svc.AssignCropToPlot(ctx, plot.ID, crop.ID, time.Now())

	// Assert
	if err != nil {
		t.Fatalf("AssignCropToPlot failed: %v", err)
	}

	if assignment.ID == 0 {
		t.Error("Expected assignment ID to be assigned")
	}

	if assignment.PlotID != plot.ID {
		t.Errorf("Expected PlotID %d, got %d", plot.ID, assignment.PlotID)
	}

	if assignment.CropID != crop.ID {
		t.Errorf("Expected CropID %d, got %d", crop.ID, assignment.CropID)
	}

	// UnassignedDateがnilであることを確認（アクティブ）
	if assignment.UnassignedDate != nil {
		t.Error("Expected UnassignedDate to be nil for active assignment")
	}

	// 区画のステータスが "occupied" に更新されていることを確認
	updatedPlot, _ := svc.GetPlotByID(ctx, plot.ID)
	if updatedPlot.Status != "occupied" {
		t.Errorf("Expected plot status 'occupied', got '%s'", updatedPlot.Status)
	}
}

// TestAssignCropToPlot_ReplaceExisting は既存配置がある場合の置き換えをテストします。
func TestAssignCropToPlot_ReplaceExisting(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 区画を作成
	plot := &model.Plot{
		UserID: 1,
		Name:   "畑A",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	}
	_ = svc.CreatePlot(ctx, plot)

	// 2つの作物を作成
	crop1 := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	}
	_ = svc.CreateCrop(ctx, crop1)

	crop2 := &model.Crop{
		UserID:              1,
		Name:                "きゅうり",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
		Status:              "planted",
	}
	_ = svc.CreateCrop(ctx, crop2)

	// 最初の作物を配置
	assignment1, _ := svc.AssignCropToPlot(ctx, plot.ID, crop1.ID, time.Now().AddDate(0, 0, -7))

	// Act: 別の作物を配置（既存を置き換え）
	assignment2, err := svc.AssignCropToPlot(ctx, plot.ID, crop2.ID, time.Now())

	// Assert
	if err != nil {
		t.Fatalf("AssignCropToPlot failed: %v", err)
	}

	// 新しい配置がアクティブ
	if assignment2.UnassignedDate != nil {
		t.Error("Expected new assignment to be active")
	}

	// 古い配置は解除されている
	oldAssignment, _ := mockRepos.GetMockPlotAssignmentRepository().GetByID(ctx, assignment1.ID)
	if oldAssignment.UnassignedDate == nil {
		t.Error("Expected old assignment to be unassigned")
	}

	// アクティブな配置は新しいものだけ
	active, _ := svc.GetActivePlotAssignment(ctx, plot.ID)
	if active.CropID != crop2.ID {
		t.Errorf("Expected active assignment for crop2, got crop %d", active.CropID)
	}
}

// =============================================================================
// UnassignCropFromPlot テスト
// =============================================================================

// TestUnassignCropFromPlot_Success は配置解除の正常動作をテストします。
func TestUnassignCropFromPlot_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 区画を作成
	plot := &model.Plot{
		UserID: 1,
		Name:   "畑A",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	}
	_ = svc.CreatePlot(ctx, plot)

	// 作物を作成
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 配置
	_, _ = svc.AssignCropToPlot(ctx, plot.ID, crop.ID, time.Now())

	// Act: 配置解除
	err := svc.UnassignCropFromPlot(ctx, plot.ID)

	// Assert
	if err != nil {
		t.Fatalf("UnassignCropFromPlot failed: %v", err)
	}

	// 区画のステータスが "available" に更新されていることを確認
	updatedPlot, _ := svc.GetPlotByID(ctx, plot.ID)
	if updatedPlot.Status != "available" {
		t.Errorf("Expected plot status 'available', got '%s'", updatedPlot.Status)
	}

	// アクティブな配置がないことを確認
	_, err = svc.GetActivePlotAssignment(ctx, plot.ID)
	if err == nil {
		t.Error("Expected no active assignment after unassign")
	}
}

// =============================================================================
// GetPlotAssignments テスト
// =============================================================================

// TestGetPlotAssignments_Success は配置履歴取得をテストします。
func TestGetPlotAssignments_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 区画を作成
	plot := &model.Plot{
		UserID: 1,
		Name:   "畑A",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	}
	_ = svc.CreatePlot(ctx, plot)

	// 複数の作物を作成し配置（履歴として）
	for i := 0; i < 3; i++ {
		crop := &model.Crop{
			UserID:              1,
			Name:                "作物" + string(rune('A'+i)),
			PlantedDate:         time.Now(),
			ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
			Status:              "planted",
		}
		_ = svc.CreateCrop(ctx, crop)
		_, _ = svc.AssignCropToPlot(ctx, plot.ID, crop.ID, time.Now().AddDate(0, 0, -i*30))
	}

	// Act: 配置履歴を取得
	assignments, err := svc.GetPlotAssignments(ctx, plot.ID)

	// Assert
	if err != nil {
		t.Fatalf("GetPlotAssignments failed: %v", err)
	}

	if len(assignments) != 3 {
		t.Errorf("Expected 3 assignments, got %d", len(assignments))
	}
}

// =============================================================================
// GetPlotLayout テスト
// =============================================================================

// TestGetPlotLayout_Success はレイアウト取得をテストします。
func TestGetPlotLayout_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 複数の区画を作成
	plot1 := &model.Plot{
		UserID: userID,
		Name:   "畑A",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	}
	_ = svc.CreatePlot(ctx, plot1)

	plot2 := &model.Plot{
		UserID: userID,
		Name:   "畑B",
		Width:  1.5,
		Height: 2.5,
		Status: "available",
	}
	_ = svc.CreatePlot(ctx, plot2)

	// 1つの区画に作物を配置
	crop := &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	}
	_ = svc.CreateCrop(ctx, crop)
	_, _ = svc.AssignCropToPlot(ctx, plot1.ID, crop.ID, time.Now())

	// Act: レイアウトを取得
	layout, err := svc.GetPlotLayout(ctx, userID)

	// Assert
	if err != nil {
		t.Fatalf("GetPlotLayout failed: %v", err)
	}

	if len(layout) != 2 {
		t.Errorf("Expected 2 layout items, got %d", len(layout))
	}

	// 配置されている区画を確認
	var assignedPlot *PlotLayoutItem
	for i := range layout {
		if layout[i].Plot.ID == plot1.ID {
			assignedPlot = &layout[i]
			break
		}
	}

	if assignedPlot == nil {
		t.Fatal("Could not find plot1 in layout")
	}

	if assignedPlot.ActiveAssignment == nil {
		t.Error("Expected plot1 to have an active assignment")
	}

	if assignedPlot.ActiveCrop == nil {
		t.Error("Expected plot1 to have an active crop")
	}

	if assignedPlot.ActiveCrop != nil && assignedPlot.ActiveCrop.Name != "トマト" {
		t.Errorf("Expected crop name 'トマト', got '%s'", assignedPlot.ActiveCrop.Name)
	}
}

// TestGetPlotLayout_Empty はユーザーに区画がない場合をテストします。
func TestGetPlotLayout_Empty(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	layout, err := svc.GetPlotLayout(ctx, 999)

	if err != nil {
		t.Fatalf("GetPlotLayout failed: %v", err)
	}

	if len(layout) != 0 {
		t.Errorf("Expected 0 layout items, got %d", len(layout))
	}
}

// =============================================================================
// GetPlotHistory テスト
// =============================================================================

// TestGetPlotHistory_Success は区画履歴取得をテストします。
func TestGetPlotHistory_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 区画を作成
	plot := &model.Plot{
		UserID: 1,
		Name:   "畑A",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	}
	_ = svc.CreatePlot(ctx, plot)

	// 複数の作物を作成し配置
	cropNames := []string{"トマト", "きゅうり", "なす"}
	for _, name := range cropNames {
		crop := &model.Crop{
			UserID:              1,
			Name:                name,
			PlantedDate:         time.Now(),
			ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
			Status:              "planted",
		}
		_ = svc.CreateCrop(ctx, crop)
		_, _ = svc.AssignCropToPlot(ctx, plot.ID, crop.ID, time.Now())
	}

	// Act: 履歴を取得
	history, err := svc.GetPlotHistory(ctx, plot.ID)

	// Assert
	if err != nil {
		t.Fatalf("GetPlotHistory failed: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("Expected 3 history items, got %d", len(history))
	}

	// 各履歴に作物情報が含まれていることを確認
	for _, item := range history {
		if item.Crop == nil {
			t.Error("Expected history item to have crop info")
		}
	}
}

// TestGetPlotHistory_Empty は履歴がない場合をテストします。
func TestGetPlotHistory_Empty(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 区画を作成（配置なし）
	plot := &model.Plot{
		UserID: 1,
		Name:   "畑A",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	}
	_ = svc.CreatePlot(ctx, plot)

	history, err := svc.GetPlotHistory(ctx, plot.ID)

	if err != nil {
		t.Fatalf("GetPlotHistory failed: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected 0 history items, got %d", len(history))
	}
}

// =============================================================================
// データ分離テスト
// =============================================================================

// TestPlotDataIsolation_DifferentUsers は異なるユーザー間のデータ分離をテストします。
func TestPlotDataIsolation_DifferentUsers(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// ユーザー1の区画
	_ = svc.CreatePlot(ctx, &model.Plot{
		UserID: 1,
		Name:   "畑A",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	})
	_ = svc.CreatePlot(ctx, &model.Plot{
		UserID: 1,
		Name:   "畑B",
		Width:  1.5,
		Height: 2.5,
		Status: "occupied",
	})

	// ユーザー2の区画
	_ = svc.CreatePlot(ctx, &model.Plot{
		UserID: 2,
		Name:   "畑C",
		Width:  3.0,
		Height: 3.0,
		Status: "available",
	})

	// Act: 各ユーザーの区画を取得
	user1Plots, _ := svc.GetUserPlots(ctx, 1)
	user2Plots, _ := svc.GetUserPlots(ctx, 2)

	// Assert: ユーザー1は2つ、ユーザー2は1つ
	if len(user1Plots) != 2 {
		t.Errorf("User 1 should have 2 plots, got %d", len(user1Plots))
	}
	if len(user2Plots) != 1 {
		t.Errorf("User 2 should have 1 plot, got %d", len(user2Plots))
	}

	// ユーザー1の区画にユーザー2のデータが含まれていないことを確認
	for _, plot := range user1Plots {
		if plot.UserID != 1 {
			t.Errorf("User 1's plots contain data from user %d", plot.UserID)
		}
	}
}

// TestPlotLayoutDataIsolation_DifferentUsers はレイアウト取得のデータ分離をテストします。
func TestPlotLayoutDataIsolation_DifferentUsers(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// ユーザー1の区画
	_ = svc.CreatePlot(ctx, &model.Plot{
		UserID: 1,
		Name:   "ユーザー1の畑",
		Width:  2.0,
		Height: 3.0,
		Status: "available",
	})

	// ユーザー2の区画
	_ = svc.CreatePlot(ctx, &model.Plot{
		UserID: 2,
		Name:   "ユーザー2の畑",
		Width:  3.0,
		Height: 3.0,
		Status: "available",
	})

	// Act: 各ユーザーのレイアウトを取得
	layout1, _ := svc.GetPlotLayout(ctx, 1)
	layout2, _ := svc.GetPlotLayout(ctx, 2)

	// Assert: 各ユーザーは自分の区画のみ取得
	if len(layout1) != 1 {
		t.Errorf("User 1 should have 1 layout item, got %d", len(layout1))
	}
	if len(layout2) != 1 {
		t.Errorf("User 2 should have 1 layout item, got %d", len(layout2))
	}

	if layout1[0].Plot.Name != "ユーザー1の畑" {
		t.Errorf("User 1's layout has wrong plot: %s", layout1[0].Plot.Name)
	}
	if layout2[0].Plot.Name != "ユーザー2の畑" {
		t.Errorf("User 2's layout has wrong plot: %s", layout2[0].Plot.Name)
	}
}
