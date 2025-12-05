// Package service - AnalyticsService Unit Tests
//
// AnalyticsServiceのユニットテストを提供します。
// MockRepositoryを使用して、データベースなしでサービス層のロジックをテストします。
//
// テスト対象:
//   - 収穫量集計（GetHarvestSummary）
//   - グラフデータ生成（GetChartData）
//   - CSVエクスポート（ExportCSV）
package service

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
)

// =============================================================================
// GetHarvestSummary テスト
// =============================================================================

// TestGetHarvestSummary_Success は収穫量集計の正常取得をテストします。
func TestGetHarvestSummary_Success(t *testing.T) {
	// Arrange: モックリポジトリとサービスをセットアップ
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 作物を作成
	crop1 := &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		PlantedDate:         time.Now().AddDate(0, -3, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop1)

	crop2 := &model.Crop{
		UserID:              userID,
		Name:                "きゅうり",
		PlantedDate:         time.Now().AddDate(0, -2, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop2)

	// 収穫データを追加（HarvestsByUserIDに直接追加）
	harvestRepo := mockRepos.GetMockHarvestRepository()
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop1.ID,
		HarvestDate:  time.Now().AddDate(0, 0, -7),
		Quantity:     2.5,
		QuantityUnit: "kg",
		Quality:      "excellent",
	})
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop1.ID,
		HarvestDate:  time.Now(),
		Quantity:     3.0,
		QuantityUnit: "kg",
		Quality:      "good",
	})
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop2.ID,
		HarvestDate:  time.Now(),
		Quantity:     1.5,
		QuantityUnit: "kg",
		Quality:      "excellent",
	})

	// Act: 収穫量集計を取得
	filter := HarvestFilter{}
	summary, err := svc.GetHarvestSummary(ctx, userID, filter)

	// Assert
	if err != nil {
		t.Fatalf("GetHarvestSummary failed: %v", err)
	}

	if summary.TotalHarvests != 3 {
		t.Errorf("Expected 3 total harvests, got %d", summary.TotalHarvests)
	}

	if summary.TotalQuantityKg != 7.0 { // 2.5 + 3.0 + 1.5
		t.Errorf("Expected 7.0 kg total, got %.2f", summary.TotalQuantityKg)
	}

	if len(summary.CropSummaries) != 2 {
		t.Errorf("Expected 2 crop summaries, got %d", len(summary.CropSummaries))
	}

	// 品質分布の確認
	if summary.QualityDistribution["excellent"] != 2 {
		t.Errorf("Expected 2 excellent quality, got %d", summary.QualityDistribution["excellent"])
	}
	if summary.QualityDistribution["good"] != 1 {
		t.Errorf("Expected 1 good quality, got %d", summary.QualityDistribution["good"])
	}
}

// TestGetHarvestSummary_WithDateFilter は日付フィルターでの収穫量集計をテストします。
func TestGetHarvestSummary_WithDateFilter(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 作物を作成
	crop := &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		PlantedDate:         time.Now().AddDate(0, -3, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 異なる日付の収穫データを追加
	harvestRepo := mockRepos.GetMockHarvestRepository()
	// 1週間前の収穫
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now().AddDate(0, 0, -7),
		Quantity:     2.0,
		QuantityUnit: "kg",
	})
	// 今日の収穫
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now(),
		Quantity:     3.0,
		QuantityUnit: "kg",
	})

	// Act: 3日前から今日までのフィルターで取得
	startDate := time.Now().AddDate(0, 0, -3)
	endDate := time.Now().Add(24 * time.Hour)
	filter := HarvestFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	summary, err := svc.GetHarvestSummary(ctx, userID, filter)

	// Assert
	if err != nil {
		t.Fatalf("GetHarvestSummary failed: %v", err)
	}

	// 3日前から今日の間には今日の収穫（3.0kg）のみが含まれる
	if summary.TotalHarvests != 1 {
		t.Errorf("Expected 1 harvest in date range, got %d", summary.TotalHarvests)
	}

	if summary.TotalQuantityKg != 3.0 {
		t.Errorf("Expected 3.0 kg in date range, got %.2f", summary.TotalQuantityKg)
	}
}

// TestGetHarvestSummary_WithCropIDFilter は作物IDフィルターでの収穫量集計をテストします。
func TestGetHarvestSummary_WithCropIDFilter(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 2つの作物を作成
	crop1 := &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		PlantedDate:         time.Now().AddDate(0, -3, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop1)

	crop2 := &model.Crop{
		UserID:              userID,
		Name:                "きゅうり",
		PlantedDate:         time.Now().AddDate(0, -2, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop2)

	// 両方の作物に収穫データを追加
	harvestRepo := mockRepos.GetMockHarvestRepository()
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop1.ID,
		HarvestDate:  time.Now(),
		Quantity:     5.0,
		QuantityUnit: "kg",
	})
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop2.ID,
		HarvestDate:  time.Now(),
		Quantity:     2.0,
		QuantityUnit: "kg",
	})

	// Act: トマトのみのフィルターで取得
	filter := HarvestFilter{
		CropID: &crop1.ID,
	}
	summary, err := svc.GetHarvestSummary(ctx, userID, filter)

	// Assert
	if err != nil {
		t.Fatalf("GetHarvestSummary failed: %v", err)
	}

	if summary.TotalHarvests != 1 {
		t.Errorf("Expected 1 harvest for crop1, got %d", summary.TotalHarvests)
	}

	if summary.TotalQuantityKg != 5.0 {
		t.Errorf("Expected 5.0 kg for crop1, got %.2f", summary.TotalQuantityKg)
	}

	if len(summary.CropSummaries) != 1 {
		t.Errorf("Expected 1 crop summary, got %d", len(summary.CropSummaries))
	}
}

// TestGetHarvestSummary_Empty はデータがない場合の収穫量集計をテストします。
func TestGetHarvestSummary_Empty(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// データなしで取得
	filter := HarvestFilter{}
	summary, err := svc.GetHarvestSummary(ctx, 999, filter)

	// Assert
	if err != nil {
		t.Fatalf("GetHarvestSummary failed: %v", err)
	}

	if summary.TotalHarvests != 0 {
		t.Errorf("Expected 0 harvests, got %d", summary.TotalHarvests)
	}

	if summary.TotalQuantityKg != 0 {
		t.Errorf("Expected 0 kg, got %.2f", summary.TotalQuantityKg)
	}

	if len(summary.CropSummaries) != 0 {
		t.Errorf("Expected 0 crop summaries, got %d", len(summary.CropSummaries))
	}
}

// TestGetHarvestSummary_UnitConversion は単位換算が正しく行われることをテストします。
func TestGetHarvestSummary_UnitConversion(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 作物を作成
	crop := &model.Crop{
		UserID:              userID,
		Name:                "ミニトマト",
		PlantedDate:         time.Now().AddDate(0, -2, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 異なる単位の収穫データを追加
	harvestRepo := mockRepos.GetMockHarvestRepository()
	// kg単位
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now(),
		Quantity:     1.0,
		QuantityUnit: "kg",
	})
	// g単位（500g = 0.5kg）
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now(),
		Quantity:     500,
		QuantityUnit: "g",
	})
	// 個数（10個 = 1.0kg として計算）
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now(),
		Quantity:     10,
		QuantityUnit: "pieces",
	})

	// Act
	filter := HarvestFilter{}
	summary, err := svc.GetHarvestSummary(ctx, userID, filter)

	// Assert
	if err != nil {
		t.Fatalf("GetHarvestSummary failed: %v", err)
	}

	// 1.0 + 0.5 + 1.0 = 2.5kg
	expectedKg := 2.5
	if summary.TotalQuantityKg != expectedKg {
		t.Errorf("Expected %.2f kg total, got %.2f", expectedKg, summary.TotalQuantityKg)
	}
}

// =============================================================================
// GetChartData テスト
// =============================================================================

// TestGetChartData_MonthlyHarvest は月別収穫量チャートデータの取得をテストします。
func TestGetChartData_MonthlyHarvest(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 作物を作成
	crop := &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		PlantedDate:         time.Now().AddDate(0, -6, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 異なる月の収穫データを追加
	harvestRepo := mockRepos.GetMockHarvestRepository()
	// 今月
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now(),
		Quantity:     2.0,
		QuantityUnit: "kg",
	})
	// 先月
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now().AddDate(0, -1, 0),
		Quantity:     3.0,
		QuantityUnit: "kg",
	})

	// Act
	filter := ChartFilter{}
	chartData, err := svc.GetChartData(ctx, userID, ChartTypeMonthlyHarvest, filter)

	// Assert
	if err != nil {
		t.Fatalf("GetChartData failed: %v", err)
	}

	if chartData.ChartType != ChartTypeMonthlyHarvest {
		t.Errorf("Expected chart type %s, got %s", ChartTypeMonthlyHarvest, chartData.ChartType)
	}

	if chartData.Title != "月別収穫量" {
		t.Errorf("Expected title '月別収穫量', got '%s'", chartData.Title)
	}

	// データの確認
	monthlyData, ok := chartData.Data.([]MonthlyHarvestData)
	if !ok {
		t.Fatal("Failed to cast data to []MonthlyHarvestData")
	}

	if len(monthlyData) != 2 {
		t.Errorf("Expected 2 monthly data points, got %d", len(monthlyData))
	}
}

// TestGetChartData_CropComparison は作物別収穫量比較チャートデータの取得をテストします。
func TestGetChartData_CropComparison(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 複数の作物を作成
	crop1 := &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		PlantedDate:         time.Now().AddDate(0, -3, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop1)

	crop2 := &model.Crop{
		UserID:              userID,
		Name:                "きゅうり",
		PlantedDate:         time.Now().AddDate(0, -2, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop2)

	// 収穫データを追加
	harvestRepo := mockRepos.GetMockHarvestRepository()
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop1.ID,
		HarvestDate:  time.Now(),
		Quantity:     8.0,
		QuantityUnit: "kg",
	})
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop2.ID,
		HarvestDate:  time.Now(),
		Quantity:     2.0,
		QuantityUnit: "kg",
	})

	// Act
	filter := ChartFilter{}
	chartData, err := svc.GetChartData(ctx, userID, ChartTypeCropComparison, filter)

	// Assert
	if err != nil {
		t.Fatalf("GetChartData failed: %v", err)
	}

	if chartData.ChartType != ChartTypeCropComparison {
		t.Errorf("Expected chart type %s, got %s", ChartTypeCropComparison, chartData.ChartType)
	}

	// データの確認
	comparisonData, ok := chartData.Data.([]CropComparisonData)
	if !ok {
		t.Fatal("Failed to cast data to []CropComparisonData")
	}

	if len(comparisonData) != 2 {
		t.Errorf("Expected 2 crop comparison data points, got %d", len(comparisonData))
	}

	// トマトが最初（収穫量順）
	if comparisonData[0].CropName != "トマト" {
		t.Errorf("Expected first crop to be 'トマト', got '%s'", comparisonData[0].CropName)
	}

	// 割合の確認（トマト: 8kg / 10kg = 80%）
	if comparisonData[0].Percentage != 80.0 {
		t.Errorf("Expected トマト percentage 80.0, got %.2f", comparisonData[0].Percentage)
	}
}

// TestGetChartData_PlotProductivity は区画生産性チャートデータの取得をテストします。
func TestGetChartData_PlotProductivity(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 区画を作成
	plot := &model.Plot{
		UserID:   userID,
		Name:     "区画A",
		Width:    2.0,
		Height:   3.0, // 6m²
		Status:   "occupied",
	}
	_ = svc.CreatePlot(ctx, plot)

	// 作物を作成
	crop := &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		PlantedDate:         time.Now().AddDate(0, -3, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 区画に作物を配置（PlotAssignmentを作成）
	_, _ = svc.AssignCropToPlot(ctx, plot.ID, crop.ID, time.Now().AddDate(0, -3, 0))

	// 収穫データを追加
	harvestRepo := mockRepos.GetMockHarvestRepository()
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now(),
		Quantity:     6.0,
		QuantityUnit: "kg",
	})

	// Act
	filter := ChartFilter{}
	chartData, err := svc.GetChartData(ctx, userID, ChartTypePlotProductivity, filter)

	// Assert
	if err != nil {
		t.Fatalf("GetChartData failed: %v", err)
	}

	if chartData.ChartType != ChartTypePlotProductivity {
		t.Errorf("Expected chart type %s, got %s", ChartTypePlotProductivity, chartData.ChartType)
	}

	// データの確認
	productivityData, ok := chartData.Data.([]PlotProductivityData)
	if !ok {
		t.Fatal("Failed to cast data to []PlotProductivityData")
	}

	if len(productivityData) != 1 {
		t.Errorf("Expected 1 plot productivity data point, got %d", len(productivityData))
	}

	// 面積あたり収穫量の確認（6kg / 6m² = 1.0 kg/m²）
	if productivityData[0].KgPerM2 != 1.0 {
		t.Errorf("Expected kg/m² 1.0, got %.2f", productivityData[0].KgPerM2)
	}
}

// TestGetChartData_InvalidType は無効なチャートタイプでエラーが返されることをテストします。
func TestGetChartData_InvalidType(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	filter := ChartFilter{}
	_, err := svc.GetChartData(ctx, 1, ChartType("invalid_type"), filter)

	if err == nil {
		t.Error("Expected error for invalid chart type, got nil")
	}
}

// TestGetChartData_Empty はデータがない場合のチャートデータ取得をテストします。
func TestGetChartData_Empty(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// データなしで取得
	filter := ChartFilter{}
	chartData, err := svc.GetChartData(ctx, 999, ChartTypeMonthlyHarvest, filter)

	// Assert
	if err != nil {
		t.Fatalf("GetChartData failed: %v", err)
	}

	// 空のデータでも正常に返される
	monthlyData, ok := chartData.Data.([]MonthlyHarvestData)
	if !ok {
		t.Fatal("Failed to cast data to []MonthlyHarvestData")
	}

	if len(monthlyData) != 0 {
		t.Errorf("Expected 0 monthly data points, got %d", len(monthlyData))
	}
}

// =============================================================================
// ExportCSV テスト
// =============================================================================

// TestExportCSV_Crops は作物データのCSVエクスポートをテストします。
func TestExportCSV_Crops(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 作物を作成
	_ = svc.CreateCrop(ctx, &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		Variety:             "桃太郎",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
		Notes:               "種から育てる",
	})
	_ = svc.CreateCrop(ctx, &model.Crop{
		UserID:              userID,
		Name:                "きゅうり",
		Variety:             "夏すずみ",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 2, 0),
		Status:              "growing",
	})

	// Act
	result, err := svc.ExportCSV(ctx, userID, ExportDataTypeCrops)

	// Assert
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	if result.DataType != ExportDataTypeCrops {
		t.Errorf("Expected data type %s, got %s", ExportDataTypeCrops, result.DataType)
	}

	if result.RecordCount != 2 {
		t.Errorf("Expected 2 records, got %d", result.RecordCount)
	}

	if result.ContentType != "text/csv; charset=utf-8" {
		t.Errorf("Expected content type 'text/csv; charset=utf-8', got '%s'", result.ContentType)
	}

	// CSVデータの確認
	csvContent := string(result.Data)
	if !strings.Contains(csvContent, "トマト") {
		t.Error("CSV should contain 'トマト'")
	}
	if !strings.Contains(csvContent, "きゅうり") {
		t.Error("CSV should contain 'きゅうり'")
	}
	if !strings.Contains(csvContent, "名前") { // ヘッダー確認
		t.Error("CSV should contain header '名前'")
	}
}

// TestExportCSV_Harvests は収穫データのCSVエクスポートをテストします。
func TestExportCSV_Harvests(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// 作物を作成
	crop := &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		PlantedDate:         time.Now().AddDate(0, -3, 0),
		ExpectedHarvestDate: time.Now(),
		Status:              "harvested",
	}
	_ = svc.CreateCrop(ctx, crop)

	// 収穫データを追加
	harvestRepo := mockRepos.GetMockHarvestRepository()
	harvestRepo.AddHarvestForUser(userID, &model.Harvest{
		CropID:       crop.ID,
		HarvestDate:  time.Now(),
		Quantity:     2.5,
		QuantityUnit: "kg",
		Quality:      "excellent",
		Notes:        "甘くて美味しい",
	})

	// Act
	result, err := svc.ExportCSV(ctx, userID, ExportDataTypeHarvests)

	// Assert
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	if result.DataType != ExportDataTypeHarvests {
		t.Errorf("Expected data type %s, got %s", ExportDataTypeHarvests, result.DataType)
	}

	if result.RecordCount != 1 {
		t.Errorf("Expected 1 record, got %d", result.RecordCount)
	}

	// CSVデータの確認
	csvContent := string(result.Data)
	if !strings.Contains(csvContent, "トマト") {
		t.Error("CSV should contain crop name 'トマト'")
	}
	if !strings.Contains(csvContent, "2.50") {
		t.Error("CSV should contain quantity '2.50'")
	}
	if !strings.Contains(csvContent, "excellent") {
		t.Error("CSV should contain quality 'excellent'")
	}
}

// TestExportCSV_Tasks はタスクデータのCSVエクスポートをテストします。
func TestExportCSV_Tasks(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// タスクを作成
	_ = svc.CreateTask(ctx, &model.Task{
		UserID:      userID,
		Title:       "水やり",
		Description: "朝と夕方に水をやる",
		DueDate:     time.Now().AddDate(0, 0, 1),
		Priority:    "high",
		Status:      "pending",
	})
	_ = svc.CreateTask(ctx, &model.Task{
		UserID:      userID,
		Title:       "肥料やり",
		Description: "週1回の肥料追加",
		DueDate:     time.Now().AddDate(0, 0, 7),
		Priority:    "medium",
		Status:      "pending",
	})

	// Act
	result, err := svc.ExportCSV(ctx, userID, ExportDataTypeTasks)

	// Assert
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	if result.DataType != ExportDataTypeTasks {
		t.Errorf("Expected data type %s, got %s", ExportDataTypeTasks, result.DataType)
	}

	if result.RecordCount != 2 {
		t.Errorf("Expected 2 records, got %d", result.RecordCount)
	}

	// CSVデータの確認
	csvContent := string(result.Data)
	if !strings.Contains(csvContent, "水やり") {
		t.Error("CSV should contain '水やり'")
	}
	if !strings.Contains(csvContent, "肥料やり") {
		t.Error("CSV should contain '肥料やり'")
	}
	if !strings.Contains(csvContent, "タイトル") { // ヘッダー確認
		t.Error("CSV should contain header 'タイトル'")
	}
}

// TestExportCSV_All は全データのZIPエクスポートをテストします。
func TestExportCSV_All(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)

	// テストデータを作成
	_ = svc.CreateCrop(ctx, &model.Crop{
		UserID:              userID,
		Name:                "トマト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 3, 0),
		Status:              "planted",
	})
	_ = svc.CreateTask(ctx, &model.Task{
		UserID:   userID,
		Title:    "水やり",
		DueDate:  time.Now().AddDate(0, 0, 1),
		Priority: "high",
		Status:   "pending",
	})

	// Act
	result, err := svc.ExportCSV(ctx, userID, ExportDataTypeAll)

	// Assert
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	if result.DataType != ExportDataTypeAll {
		t.Errorf("Expected data type %s, got %s", ExportDataTypeAll, result.DataType)
	}

	if result.ContentType != "application/zip" {
		t.Errorf("Expected content type 'application/zip', got '%s'", result.ContentType)
	}

	if !strings.HasSuffix(result.FileName, ".zip") {
		t.Errorf("Expected filename to end with '.zip', got '%s'", result.FileName)
	}

	// ZIPファイルの内容確認
	if len(result.Data) == 0 {
		t.Error("Expected non-empty ZIP data")
	}

	// ZIPファイルが正しい形式か確認
	reader := bytes.NewReader(result.Data)
	if reader.Len() < 4 {
		t.Error("ZIP file is too small")
	}

	// ZIPマジックナンバーの確認 (PK\x03\x04)
	magic := make([]byte, 4)
	_, err = reader.Read(magic)
	if err != nil {
		t.Fatalf("Failed to read ZIP magic number: %v", err)
	}
	if magic[0] != 'P' || magic[1] != 'K' {
		t.Error("Invalid ZIP file format")
	}
}

// TestExportCSV_InvalidType は無効なデータタイプでエラーが返されることをテストします。
func TestExportCSV_InvalidType(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	_, err := svc.ExportCSV(ctx, 1, ExportDataType("invalid_type"))

	if err == nil {
		t.Error("Expected error for invalid data type, got nil")
	}
}

// TestExportCSV_Empty はデータがない場合のCSVエクスポートをテストします。
func TestExportCSV_Empty(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// データなしでエクスポート
	result, err := svc.ExportCSV(ctx, 999, ExportDataTypeCrops)

	// Assert
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	if result.RecordCount != 0 {
		t.Errorf("Expected 0 records, got %d", result.RecordCount)
	}

	// 空でもヘッダーは含まれる
	csvContent := string(result.Data)
	if !strings.Contains(csvContent, "名前") {
		t.Error("CSV should contain header even when empty")
	}
}

// TestExportCSV_BOMPresent はCSVにBOM（Byte Order Mark）が含まれることをテストします。
func TestExportCSV_BOMPresent(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	userID := uint(1)
	_ = svc.CreateCrop(ctx, &model.Crop{
		UserID:              userID,
		Name:                "テスト",
		PlantedDate:         time.Now(),
		ExpectedHarvestDate: time.Now().AddDate(0, 1, 0),
		Status:              "planted",
	})

	// Act
	result, err := svc.ExportCSV(ctx, userID, ExportDataTypeCrops)

	// Assert
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	// BOMの確認（UTF-8 BOM: 0xEF 0xBB 0xBF）
	if len(result.Data) < 3 {
		t.Fatal("CSV data is too short")
	}

	if result.Data[0] != 0xEF || result.Data[1] != 0xBB || result.Data[2] != 0xBF {
		t.Error("CSV should start with UTF-8 BOM for Excel compatibility")
	}
}
