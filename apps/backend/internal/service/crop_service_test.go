// Package service - CropService Unit Tests
//
// CropServiceのユニットテストを提供します。
// MockRepositoryを使用して、データベースなしでサービス層のロジックをテストします。
//
// テスト対象:
//   - 作物CRUD操作
//   - 成長記録操作
//   - 収穫記録操作
//   - データ分離（ユーザーIDによるフィルタリング）
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
)

// =============================================================================
// CreateCrop テスト
// =============================================================================

// TestCreateCrop_Success は作物の正常作成をテストします。
func TestCreateCrop_Success(t *testing.T) {
	// Arrange: モックリポジトリとサービスをセットアップ
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// テスト用の作物データ
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		Variety:             "桃太郎",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0), // 3ヶ月後
		Status:              "planted",
		Notes:               "種から育てる",
	}

	// Act: 作物を作成
	err := svc.CreateCrop(ctx, crop)

	// Assert: エラーがないことを確認
	if err != nil {
		t.Fatalf("CreateCrop failed: %v", err)
	}

	// IDが割り当てられていることを確認
	if crop.ID == 0 {
		t.Error("Expected crop ID to be assigned, got 0")
	}

	// モックに保存されていることを確認
	savedCrop, err := mockRepos.GetMockCropRepository().GetByID(ctx, crop.ID)
	if err != nil {
		t.Fatalf("Failed to get saved crop: %v", err)
	}

	if savedCrop.Name != "トマト" {
		t.Errorf("Expected name 'トマト', got '%s'", savedCrop.Name)
	}
}

// TestCreateCrop_WithPlotID は区画ID付きの作物作成をテストします。
func TestCreateCrop_WithPlotID(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	plotID := uint(5)
	crop := &model.Crop{
		UserID:              1,
		PlotID:              &plotID,
		Name:                "きゅうり",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
		Status:              "planted",
	}

	err := svc.CreateCrop(ctx, crop)

	if err != nil {
		t.Fatalf("CreateCrop failed: %v", err)
	}

	if crop.PlotID == nil || *crop.PlotID != 5 {
		t.Error("Expected PlotID to be 5")
	}
}

// TestCreateCrop_Error はリポジトリエラー時の動作をテストします。
func TestCreateCrop_Error(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// エラーを返すように設定
	mockRepos.GetMockCropRepository().CreateFunc = func(ctx context.Context, crop *model.Crop) error {
		return errors.New("database error")
	}

	crop := &model.Crop{
		UserID:              1,
		Name:                "なす",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
	}

	err := svc.CreateCrop(ctx, crop)

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// =============================================================================
// GetUserCrops テスト
// =============================================================================

// TestGetUserCrops_Success はユーザーの作物一覧取得をテストします。
func TestGetUserCrops_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// テストデータを準備
	userID := uint(1)
	crops := []*model.Crop{
		{
			UserID:              userID,
			Name:                "トマト",
			PlantedDate:         time.Now().AddDate(0, -1, 0),
			ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
			Status:              "growing",
		},
		{
			UserID:              userID,
			Name:                "きゅうり",
			PlantedDate:         time.Now().AddDate(0, -2, 0),
			ExpectedHarvestDate: time.Now().AddDate(0, 1, 0),
			Status:              "ready_to_harvest",
		},
	}

	// モックに作物を追加
	for _, crop := range crops {
		_ = svc.CreateCrop(ctx, crop)
	}

	// Act: ユーザーの作物を取得
	result, err := svc.GetUserCrops(ctx, userID)

	// Assert
	if err != nil {
		t.Fatalf("GetUserCrops failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 crops, got %d", len(result))
	}
}

// TestGetUserCrops_Empty はユーザーに作物がない場合をテストします。
func TestGetUserCrops_Empty(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	result, err := svc.GetUserCrops(ctx, 999) // 存在しないユーザー

	if err != nil {
		t.Fatalf("GetUserCrops failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 crops, got %d", len(result))
	}
}

// =============================================================================
// GetUserCropsByStatus テスト
// =============================================================================

// TestGetUserCropsByStatus_Success はステータスでフィルタリングした作物取得をテストします。
func TestGetUserCropsByStatus_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 異なるステータスの作物を作成
	_ = svc.CreateCrop(ctx, &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	})
	_ = svc.CreateCrop(ctx, &model.Crop{
		UserID:              userID,
		Name:                "きゅうり",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
		Status:              "growing",
	})
	_ = svc.CreateCrop(ctx, &model.Crop{
		UserID:              userID,
		Name:                "なす",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
		Status:              "growing",
	})

	// Act: "growing" ステータスでフィルタ
	result, err := svc.GetUserCropsByStatus(ctx, userID, "growing")

	// Assert
	if err != nil {
		t.Fatalf("GetUserCropsByStatus failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 growing crops, got %d", len(result))
	}

	for _, crop := range result {
		if crop.Status != "growing" {
			t.Errorf("Expected status 'growing', got '%s'", crop.Status)
		}
	}
}

// =============================================================================
// UpdateCrop テスト
// =============================================================================

// TestUpdateCrop_Success は作物の更新をテストします。
func TestUpdateCrop_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 作物を作成
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	}
	_ = svc.CreateCrop(ctx, crop)

	// Act: ステータスを更新
	crop.Status = "growing"
	crop.Notes = "順調に成長中"
	err := svc.UpdateCrop(ctx, crop)

	// Assert
	if err != nil {
		t.Fatalf("UpdateCrop failed: %v", err)
	}

	// 更新が反映されていることを確認
	updated, _ := svc.GetCropByID(ctx, crop.ID)
	if updated.Status != "growing" {
		t.Errorf("Expected status 'growing', got '%s'", updated.Status)
	}
	if updated.Notes != "順調に成長中" {
		t.Errorf("Expected notes '順調に成長中', got '%s'", updated.Notes)
	}
}

// =============================================================================
// DeleteCrop テスト
// =============================================================================

// TestDeleteCrop_Success は作物の削除をテストします。
func TestDeleteCrop_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 作物を作成
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	}
	_ = svc.CreateCrop(ctx, crop)

	// Act: 削除
	err := svc.DeleteCrop(ctx, crop.ID)

	// Assert
	if err != nil {
		t.Fatalf("DeleteCrop failed: %v", err)
	}

	// 削除されていることを確認
	_, err = svc.GetCropByID(ctx, crop.ID)
	if err == nil {
		t.Error("Expected crop to be deleted, but it still exists")
	}
}

// TestDeleteCrop_WithRelatedRecords は関連レコードも含めて削除されることをテストします。
func TestDeleteCrop_WithRelatedRecords(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 作物を作成
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "growing",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 成長記録を追加
	growthRecord := &model.GrowthRecord{
		CropID:      crop.ID,
		RecordDate:  time.Now(),
		GrowthStage: "vegetative",
		Notes:       "葉が増えてきた",
	}
	_ = svc.CreateGrowthRecord(ctx, growthRecord)

	// 収穫記録を追加
	harvest := &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now(),
		Quantity:     5.0,
		QuantityUnit: "kg",
	}
	_ = svc.CreateHarvest(ctx, harvest)

	// Act: 作物を削除（関連レコードも削除される）
	err := svc.DeleteCrop(ctx, crop.ID)

	// Assert
	if err != nil {
		t.Fatalf("DeleteCrop failed: %v", err)
	}

	// 成長記録も削除されていることを確認
	records, _ := svc.GetCropGrowthRecords(ctx, crop.ID)
	if len(records) != 0 {
		t.Errorf("Expected 0 growth records after deletion, got %d", len(records))
	}

	// 収穫記録も削除されていることを確認
	harvests, _ := svc.GetCropHarvests(ctx, crop.ID)
	if len(harvests) != 0 {
		t.Errorf("Expected 0 harvests after deletion, got %d", len(harvests))
	}
}

// =============================================================================
// GrowthRecord テスト
// =============================================================================

// TestCreateGrowthRecord_Success は成長記録の正常作成をテストします。
func TestCreateGrowthRecord_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 先に作物を作成
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "growing",
	}
	_ = svc.CreateCrop(ctx, crop)

	// Act: 成長記録を作成
	record := &model.GrowthRecord{
		CropID:      crop.ID,
		RecordDate:  time.Now(),
		GrowthStage: "seedling",
		Notes:       "双葉が出た",
	}
	err := svc.CreateGrowthRecord(ctx, record)

	// Assert
	if err != nil {
		t.Fatalf("CreateGrowthRecord failed: %v", err)
	}

	if record.ID == 0 {
		t.Error("Expected growth record ID to be assigned")
	}
}

// TestGetCropGrowthRecords_Success は作物の成長記録一覧取得をテストします。
func TestGetCropGrowthRecords_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 作物を作成
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "growing",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 複数の成長記録を追加
	stages := []string{"seedling", "vegetative", "flowering"}
	for i, stage := range stages {
		_ = svc.CreateGrowthRecord(ctx, &model.GrowthRecord{
			CropID:      crop.ID,
			RecordDate:  time.Now().AddDate(0, 0, i*14), // 2週間ごと
			GrowthStage: stage,
		})
	}

	// Act: 成長記録を取得
	records, err := svc.GetCropGrowthRecords(ctx, crop.ID)

	// Assert
	if err != nil {
		t.Fatalf("GetCropGrowthRecords failed: %v", err)
	}

	if len(records) != 3 {
		t.Errorf("Expected 3 growth records, got %d", len(records))
	}
}

// TestCreateGrowthRecord_AllStages は全ての成長段階をテストします。
func TestCreateGrowthRecord_AllStages(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "growing",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 全ての有効な成長段階
	validStages := []string{"seedling", "vegetative", "flowering", "fruiting"}

	for _, stage := range validStages {
		record := &model.GrowthRecord{
			CropID:      crop.ID,
			RecordDate:  time.Now(),
			GrowthStage: stage,
		}
		err := svc.CreateGrowthRecord(ctx, record)
		if err != nil {
			t.Errorf("Failed to create growth record with stage '%s': %v", stage, err)
		}
	}
}

// =============================================================================
// Harvest テスト
// =============================================================================

// TestCreateHarvest_Success は収穫記録の正常作成をテストします。
func TestCreateHarvest_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 作物を作成
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now().AddDate(0, -3, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "ready_to_harvest",
	}
	_ = svc.CreateCrop(ctx, crop)

	// Act: 収穫記録を作成
	harvest := &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now(),
		Quantity:     2.5,
		QuantityUnit: "kg",
		Quality:      "excellent",
		Notes:        "甘くて美味しい",
	}
	err := svc.CreateHarvest(ctx, harvest)

	// Assert
	if err != nil {
		t.Fatalf("CreateHarvest failed: %v", err)
	}

	if harvest.ID == 0 {
		t.Error("Expected harvest ID to be assigned")
	}
}

// TestGetCropHarvests_Success は作物の収穫記録一覧取得をテストします。
func TestGetCropHarvests_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 作物を作成
	crop := &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now().AddDate(0, -3, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 複数の収穫記録を追加（複数回収穫可能な作物）
	for i := 0; i < 3; i++ {
		_ = svc.CreateHarvest(ctx, &model.Harvest{
			CropID:       crop.ID,
			HarvestDate:  time.Now().AddDate(0, 0, i*7),
			Quantity:     float64(i+1) * 0.5,
			QuantityUnit: "kg",
		})
	}

	// Act: 収穫記録を取得
	harvests, err := svc.GetCropHarvests(ctx, crop.ID)

	// Assert
	if err != nil {
		t.Fatalf("GetCropHarvests failed: %v", err)
	}

	if len(harvests) != 3 {
		t.Errorf("Expected 3 harvest records, got %d", len(harvests))
	}
}

// TestCreateHarvest_AllQuantityUnits は全ての数量単位をテストします。
func TestCreateHarvest_AllQuantityUnits(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	crop := &model.Crop{
		UserID:              1,
		Name:                "野菜",
		PlantedDate:         time.Now().AddDate(0, -2, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 全ての有効な数量単位
	validUnits := []string{"kg", "g", "pieces"}

	for _, unit := range validUnits {
		harvest := &model.Harvest{
			CropID:       crop.ID,
			HarvestDate:  time.Now(),
			Quantity:     10.0,
			QuantityUnit: unit,
		}
		err := svc.CreateHarvest(ctx, harvest)
		if err != nil {
			t.Errorf("Failed to create harvest with unit '%s': %v", unit, err)
		}
	}
}

// TestCreateHarvest_AllQualityLevels は全ての品質レベルをテストします。
func TestCreateHarvest_AllQualityLevels(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	crop := &model.Crop{
		UserID:              1,
		Name:                "野菜",
		PlantedDate:         time.Now().AddDate(0, -2, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 全ての有効な品質レベル
	validQualities := []string{"excellent", "good", "fair", "poor"}

	for _, quality := range validQualities {
		harvest := &model.Harvest{
			CropID:       crop.ID,
			HarvestDate:  time.Now(),
			Quantity:     1.0,
			QuantityUnit: "kg",
			Quality:      quality,
		}
		err := svc.CreateHarvest(ctx, harvest)
		if err != nil {
			t.Errorf("Failed to create harvest with quality '%s': %v", quality, err)
		}
	}
}

// =============================================================================
// データ分離テスト
// =============================================================================

// TestDataIsolation_DifferentUsers は異なるユーザー間のデータ分離をテストします。
func TestDataIsolation_DifferentUsers(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// ユーザー1の作物
	_ = svc.CreateCrop(ctx, &model.Crop{
		UserID:              1,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	})
	_ = svc.CreateCrop(ctx, &model.Crop{
		UserID:              1,
		Name:                "きゅうり",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
		Status:              "planted",
	})

	// ユーザー2の作物
	_ = svc.CreateCrop(ctx, &model.Crop{
		UserID:              2,
		Name:                "なす",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	})

	// Act: 各ユーザーの作物を取得
	user1Crops, _ := svc.GetUserCrops(ctx, 1)
	user2Crops, _ := svc.GetUserCrops(ctx, 2)

	// Assert: ユーザー1は2つ、ユーザー2は1つ
	if len(user1Crops) != 2 {
		t.Errorf("User 1 should have 2 crops, got %d", len(user1Crops))
	}
	if len(user2Crops) != 1 {
		t.Errorf("User 2 should have 1 crop, got %d", len(user2Crops))
	}

	// ユーザー1の作物にユーザー2のデータが含まれていないことを確認
	for _, crop := range user1Crops {
		if crop.UserID != 1 {
			t.Errorf("User 1's crops contain data from user %d", crop.UserID)
		}
	}
}
