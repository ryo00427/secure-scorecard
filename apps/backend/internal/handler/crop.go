// Package handler - Crop Handler
//
// 作物管理のHTTPハンドラを提供します。
// エンドポイント:
//   - GET    /api/v1/crops           - ユーザーの全作物取得
//   - GET    /api/v1/crops/:id       - 特定の作物取得
//   - POST   /api/v1/crops           - 新規作物登録
//   - PUT    /api/v1/crops/:id       - 作物更新
//   - DELETE /api/v1/crops/:id       - 作物削除
//   - POST   /api/v1/crops/:id/growth-records - 成長記録追加
//   - GET    /api/v1/crops/:id/growth-records - 成長記録一覧取得
//   - POST   /api/v1/crops/:id/harvests       - 収穫記録追加
//   - GET    /api/v1/crops/:id/harvests       - 収穫記録一覧取得
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/auth"
	apperrors "github.com/secure-scorecard/backend/internal/errors"
	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/validator"
)

// =============================================================================
// Request/Response 構造体
// =============================================================================

// CreateCropRequest は作物登録リクエストの構造体です。
//
// フィールド:
//   - Name: 作物名（必須、最大100文字）
//   - Variety: 品種（任意、最大100文字）
//   - PlantedDate: 植え付け日（必須）
//   - ExpectedHarvestDate: 予想収穫日（必須）
//   - PlotID: 区画ID（任意）
//   - Notes: メモ（任意、最大1000文字）
type CreateCropRequest struct {
	Name                string    `json:"name" validate:"required,max=100"`
	Variety             string    `json:"variety" validate:"max=100"`
	PlantedDate         time.Time `json:"planted_date" validate:"required"`
	ExpectedHarvestDate time.Time `json:"expected_harvest_date" validate:"required"`
	PlotID              *uint     `json:"plot_id"`
	Notes               string    `json:"notes" validate:"max=1000"`
}

// UpdateCropRequest は作物更新リクエストの構造体です。
// すべてのフィールドは任意で、指定されたフィールドのみ更新されます。
type UpdateCropRequest struct {
	Name                string    `json:"name" validate:"max=100"`
	Variety             string    `json:"variety" validate:"max=100"`
	PlantedDate         time.Time `json:"planted_date"`
	ExpectedHarvestDate time.Time `json:"expected_harvest_date"`
	Status              string    `json:"status" validate:"omitempty,oneof=planted growing ready_to_harvest harvested failed"`
	PlotID              *uint     `json:"plot_id"`
	Notes               string    `json:"notes" validate:"max=1000"`
}

// CreateGrowthRecordRequest は成長記録追加リクエストの構造体です。
//
// 成長段階:
//   - seedling: 苗
//   - vegetative: 成長期
//   - flowering: 開花期
//   - fruiting: 結実期
type CreateGrowthRecordRequest struct {
	RecordDate  time.Time `json:"record_date" validate:"required"`
	GrowthStage string    `json:"growth_stage" validate:"required,oneof=seedling vegetative flowering fruiting"`
	Notes       string    `json:"notes" validate:"max=1000"`
	ImageURL    string    `json:"image_url" validate:"max=500"`
}

// CreateHarvestRequest は収穫記録追加リクエストの構造体です。
//
// 品質評価:
//   - excellent: 優良
//   - good: 良好
//   - fair: 普通
//   - poor: 不良
type CreateHarvestRequest struct {
	HarvestDate  time.Time `json:"harvest_date" validate:"required"`
	Quantity     float64   `json:"quantity" validate:"required,gt=0"`
	QuantityUnit string    `json:"quantity_unit" validate:"required,oneof=kg g pieces"`
	Quality      string    `json:"quality" validate:"omitempty,oneof=excellent good fair poor"`
	Notes        string    `json:"notes" validate:"max=1000"`
}

// =============================================================================
// Crop ハンドラメソッド
// =============================================================================

// GetCrops はユーザーの全作物を取得します。
//
// クエリパラメータ:
//   - status: フィルタするステータス（planted/growing/ready_to_harvest/harvested/failed）
//
// レスポンス:
//   - 200: 作物の配列（植え付け日順）
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) GetCrops(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// statusクエリパラメータでフィルタリング
	status := c.QueryParam("status")
	var crops []model.Crop
	var err error

	if status != "" {
		// ステータスでフィルタ
		crops, err = h.service.GetUserCropsByStatus(ctx, userID, status)
	} else {
		// 全作物取得
		crops, err = h.service.GetUserCrops(ctx, userID)
	}

	if err != nil {
		return apperrors.NewInternalError("Failed to fetch crops")
	}

	return c.JSON(http.StatusOK, crops)
}

// GetCrop は特定の作物を取得します。
//
// パスパラメータ:
//   - id: 作物ID
//
// レスポンス:
//   - 200: 作物オブジェクト
//   - 400: 無効なID形式
//   - 404: 作物が見つからない
func (h *Handler) GetCrop(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid crop ID")
	}

	// 作物を取得
	crop, err := h.service.GetCropByID(ctx, uint(id))
	if err != nil {
		return apperrors.NewNotFoundError("Crop")
	}

	return c.JSON(http.StatusOK, crop)
}

// CreateCrop は新しい作物を登録します。
//
// リクエストボディ:
//   - name: 作物名（必須）
//   - variety: 品種（任意）
//   - planted_date: 植え付け日（必須）
//   - expected_harvest_date: 予想収穫日（必須）
//   - plot_id: 区画ID（任意）
//   - notes: メモ（任意）
//
// レスポンス:
//   - 201: 登録された作物
//   - 400: バリデーションエラー
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) CreateCrop(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// リクエストボディをバインド&バリデーション
	var req CreateCropRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 日付バリデーション: plantedDate <= expectedHarvestDate
	if req.PlantedDate.After(req.ExpectedHarvestDate) {
		return apperrors.NewBadRequestError("planted_date must be before or equal to expected_harvest_date")
	}

	// 作物モデルを作成
	crop := &model.Crop{
		UserID:              userID,
		PlotID:              req.PlotID,
		Name:                req.Name,
		Variety:             req.Variety,
		PlantedDate:         req.PlantedDate,
		ExpectedHarvestDate: req.ExpectedHarvestDate,
		Status:              "planted", // 新規作物は常に planted
		Notes:               req.Notes,
	}

	// DBに保存
	if err := h.service.CreateCrop(ctx, crop); err != nil {
		return apperrors.NewInternalError("Failed to create crop")
	}

	return c.JSON(http.StatusCreated, crop)
}

// UpdateCrop は既存の作物を更新します。
//
// パスパラメータ:
//   - id: 作物ID
//
// リクエストボディ: 更新するフィールド（任意）
//
// レスポンス:
//   - 200: 更新された作物
//   - 400: バリデーションエラー
//   - 404: 作物が見つからない
//   - 500: 内部エラー
func (h *Handler) UpdateCrop(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid crop ID")
	}

	// リクエストボディをバインド&バリデーション
	var req UpdateCropRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 既存の作物を取得
	crop, err := h.service.GetCropByID(ctx, uint(id))
	if err != nil {
		return apperrors.NewNotFoundError("Crop")
	}

	// リクエストで指定されたフィールドのみ更新
	if req.Name != "" {
		crop.Name = req.Name
	}
	if req.Variety != "" {
		crop.Variety = req.Variety
	}
	if !req.PlantedDate.IsZero() {
		crop.PlantedDate = req.PlantedDate
	}
	if !req.ExpectedHarvestDate.IsZero() {
		crop.ExpectedHarvestDate = req.ExpectedHarvestDate
	}
	if req.Status != "" {
		crop.Status = req.Status
	}
	if req.PlotID != nil {
		crop.PlotID = req.PlotID
	}
	if req.Notes != "" {
		crop.Notes = req.Notes
	}

	// 日付バリデーション: plantedDate <= expectedHarvestDate
	if crop.PlantedDate.After(crop.ExpectedHarvestDate) {
		return apperrors.NewBadRequestError("planted_date must be before or equal to expected_harvest_date")
	}

	// DBを更新
	if err := h.service.UpdateCrop(ctx, crop); err != nil {
		return apperrors.NewInternalError("Failed to update crop")
	}

	return c.JSON(http.StatusOK, crop)
}

// DeleteCrop は作物を削除します（論理削除）。
// 関連する成長記録と収穫記録も削除されます。
//
// パスパラメータ:
//   - id: 作物ID
//
// レスポンス:
//   - 204: 削除成功（コンテンツなし）
//   - 400: 無効なID形式
//   - 500: 内部エラー
func (h *Handler) DeleteCrop(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid crop ID")
	}

	// 作物を削除（関連データも含む）
	if err := h.service.DeleteCrop(ctx, uint(id)); err != nil {
		return apperrors.NewInternalError("Failed to delete crop")
	}

	return c.NoContent(http.StatusNoContent)
}

// =============================================================================
// GrowthRecord ハンドラメソッド
// =============================================================================

// GetGrowthRecords は作物の全成長記録を取得します。
//
// パスパラメータ:
//   - id: 作物ID
//
// レスポンス:
//   - 200: 成長記録の配列
//   - 400: 無効なID形式
//   - 500: 内部エラー
func (h *Handler) GetGrowthRecords(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	cropID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid crop ID")
	}

	// 成長記録を取得
	records, err := h.service.GetCropGrowthRecords(ctx, uint(cropID))
	if err != nil {
		return apperrors.NewInternalError("Failed to fetch growth records")
	}

	return c.JSON(http.StatusOK, records)
}

// CreateGrowthRecord は新しい成長記録を追加します。
//
// パスパラメータ:
//   - id: 作物ID
//
// リクエストボディ:
//   - record_date: 記録日（必須）
//   - growth_stage: 成長段階（必須）
//   - notes: メモ（任意）
//   - image_url: 画像URL（任意）
//
// レスポンス:
//   - 201: 追加された成長記録
//   - 400: バリデーションエラー
//   - 500: 内部エラー
func (h *Handler) CreateGrowthRecord(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	cropID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid crop ID")
	}

	// リクエストボディをバインド&バリデーション
	var req CreateGrowthRecordRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 成長記録モデルを作成
	record := &model.GrowthRecord{
		CropID:      uint(cropID),
		RecordDate:  req.RecordDate,
		GrowthStage: req.GrowthStage,
		Notes:       req.Notes,
		ImageURL:    req.ImageURL,
	}

	// DBに保存
	if err := h.service.CreateGrowthRecord(ctx, record); err != nil {
		return apperrors.NewInternalError("Failed to create growth record")
	}

	return c.JSON(http.StatusCreated, record)
}

// =============================================================================
// Harvest ハンドラメソッド
// =============================================================================

// GetHarvests は作物の全収穫記録を取得します。
//
// パスパラメータ:
//   - id: 作物ID
//
// レスポンス:
//   - 200: 収穫記録の配列
//   - 400: 無効なID形式
//   - 500: 内部エラー
func (h *Handler) GetHarvests(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	cropID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid crop ID")
	}

	// 収穫記録を取得
	harvests, err := h.service.GetCropHarvests(ctx, uint(cropID))
	if err != nil {
		return apperrors.NewInternalError("Failed to fetch harvests")
	}

	return c.JSON(http.StatusOK, harvests)
}

// CreateHarvest は新しい収穫記録を追加します。
//
// パスパラメータ:
//   - id: 作物ID
//
// リクエストボディ:
//   - harvest_date: 収穫日（必須）
//   - quantity: 収穫量（必須、0より大きい）
//   - quantity_unit: 単位（必須、kg/g/pieces）
//   - quality: 品質（任意）
//   - notes: メモ（任意）
//
// レスポンス:
//   - 201: 追加された収穫記録
//   - 400: バリデーションエラー
//   - 500: 内部エラー
func (h *Handler) CreateHarvest(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	cropID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid crop ID")
	}

	// リクエストボディをバインド&バリデーション
	var req CreateHarvestRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 収穫記録モデルを作成
	harvest := &model.Harvest{
		CropID:       uint(cropID),
		HarvestDate:  req.HarvestDate,
		Quantity:     req.Quantity,
		QuantityUnit: req.QuantityUnit,
		Quality:      req.Quality,
		Notes:        req.Notes,
	}

	// DBに保存
	if err := h.service.CreateHarvest(ctx, harvest); err != nil {
		return apperrors.NewInternalError("Failed to create harvest")
	}

	return c.JSON(http.StatusCreated, harvest)
}
