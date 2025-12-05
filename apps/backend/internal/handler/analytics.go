// Package handler - Analytics HTTP Handlers
//
// 分析データ取得のHTTPハンドラを提供します。
//
// エンドポイント:
//   - GET /api/v1/analytics/harvest - 収穫量集計取得
//   - GET /api/v1/analytics/charts/:type - グラフデータ取得
//   - GET /api/v1/analytics/export/:dataType - CSVエクスポート
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/auth"
	apperrors "github.com/secure-scorecard/backend/internal/errors"
	"github.com/secure-scorecard/backend/internal/service"
)

// =============================================================================
// Analytics ハンドラメソッド
// =============================================================================

// GetHarvestSummary は収穫量集計を取得します。
// ユーザーの収穫データを集計し、作物ごとの統計情報を返します。
//
// クエリパラメータ:
//   - start_date: 開始日（YYYY-MM-DD形式、省略可）
//   - end_date: 終了日（YYYY-MM-DD形式、省略可）
//   - crop_id: 作物ID（省略可、指定時はその作物のみ集計）
//
// レスポンス:
//   - 200: HarvestSummary オブジェクト
//   - 400: パラメータ形式エラー
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) GetHarvestSummary(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// フィルタ条件を解析
	filter := service.HarvestFilter{}

	// 開始日
	if startDateStr := c.QueryParam("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return apperrors.NewBadRequestError("Invalid start_date format. Use YYYY-MM-DD")
		}
		filter.StartDate = &startDate
	}

	// 終了日
	if endDateStr := c.QueryParam("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return apperrors.NewBadRequestError("Invalid end_date format. Use YYYY-MM-DD")
		}
		// 終了日は当日の終わりまでを含む
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.EndDate = &endDate
	}

	// 作物ID
	if cropIDStr := c.QueryParam("crop_id"); cropIDStr != "" {
		cropID, err := strconv.ParseUint(cropIDStr, 10, 32)
		if err != nil {
			return apperrors.NewBadRequestError("Invalid crop_id")
		}
		cropIDUint := uint(cropID)
		filter.CropID = &cropIDUint
	}

	// 集計を取得
	summary, err := h.service.GetHarvestSummary(ctx, userID, filter)
	if err != nil {
		return apperrors.NewInternalError("Failed to fetch harvest summary")
	}

	return c.JSON(http.StatusOK, summary)
}

// GetChartData はグラフ表示用のデータを取得します。
// グラフの種類に応じたデータを生成して返します。
//
// パスパラメータ:
//   - type: グラフの種類（monthly_harvest, crop_comparison, plot_productivity）
//
// クエリパラメータ:
//   - start_date: 開始日（YYYY-MM-DD形式、省略可）
//   - end_date: 終了日（YYYY-MM-DD形式、省略可）
//   - year: 対象年（省略可）
//
// レスポンス:
//   - 200: ChartData オブジェクト
//   - 400: パラメータ形式エラーまたは不正なグラフ種類
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) GetChartData(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// グラフ種類を取得
	chartTypeStr := c.Param("type")
	if chartTypeStr == "" {
		return apperrors.NewBadRequestError("Chart type is required")
	}

	// グラフ種類をバリデーション
	chartType := service.ChartType(chartTypeStr)
	validTypes := map[service.ChartType]bool{
		service.ChartTypeMonthlyHarvest:   true,
		service.ChartTypeCropComparison:   true,
		service.ChartTypePlotProductivity: true,
	}
	if !validTypes[chartType] {
		return apperrors.NewBadRequestError("Invalid chart type. Valid types: monthly_harvest, crop_comparison, plot_productivity")
	}

	// フィルタ条件を解析
	filter := service.ChartFilter{}

	// 開始日
	if startDateStr := c.QueryParam("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return apperrors.NewBadRequestError("Invalid start_date format. Use YYYY-MM-DD")
		}
		filter.StartDate = &startDate
	}

	// 終了日
	if endDateStr := c.QueryParam("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return apperrors.NewBadRequestError("Invalid end_date format. Use YYYY-MM-DD")
		}
		// 終了日は当日の終わりまでを含む
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.EndDate = &endDate
	}

	// 年
	if yearStr := c.QueryParam("year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			return apperrors.NewBadRequestError("Invalid year")
		}
		filter.Year = &year
	}

	// グラフデータを取得
	chartData, err := h.service.GetChartData(ctx, userID, chartType, filter)
	if err != nil {
		return apperrors.NewInternalError("Failed to generate chart data")
	}

	return c.JSON(http.StatusOK, chartData)
}

// ExportCSV はデータをCSV形式でエクスポートします。
// データ種類に応じたCSVファイルまたはZIPファイルをダウンロードとして返します。
//
// パスパラメータ:
//   - dataType: エクスポートするデータ種類（crops, harvests, tasks, all）
//
// レスポンス:
//   - 200: CSV/ZIPファイル（Content-Disposition: attachment）
//   - 400: 不正なデータ種類
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) ExportCSV(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// データ種類を取得
	dataTypeStr := c.Param("dataType")
	if dataTypeStr == "" {
		return apperrors.NewBadRequestError("Data type is required")
	}

	// データ種類をバリデーション
	dataType := service.ExportDataType(dataTypeStr)
	validTypes := map[service.ExportDataType]bool{
		service.ExportDataTypeCrops:    true,
		service.ExportDataTypeHarvests: true,
		service.ExportDataTypeTasks:    true,
		service.ExportDataTypeAll:      true,
	}
	if !validTypes[dataType] {
		return apperrors.NewBadRequestError("Invalid data type. Valid types: crops, harvests, tasks, all")
	}

	// CSVをエクスポート
	result, err := h.service.ExportCSV(ctx, userID, dataType)
	if err != nil {
		return apperrors.NewInternalError("Failed to export CSV")
	}

	// レスポンスヘッダーを設定
	c.Response().Header().Set("Content-Disposition", "attachment; filename=\""+result.FileName+"\"")
	c.Response().Header().Set("Content-Type", result.ContentType)

	return c.Blob(http.StatusOK, result.ContentType, result.Data)
}
