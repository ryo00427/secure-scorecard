// Package handler - Plot Handler
//
// 区画管理のHTTPハンドラを提供します。
// エンドポイント:
//   - GET    /api/v1/plots              - ユーザーの全区画取得
//   - GET    /api/v1/plots/:id          - 特定の区画取得
//   - POST   /api/v1/plots              - 新規区画作成
//   - PUT    /api/v1/plots/:id          - 区画更新
//   - DELETE /api/v1/plots/:id          - 区画削除
//   - POST   /api/v1/plots/:id/assign   - 作物を区画に配置
//   - DELETE /api/v1/plots/:id/assign   - 配置解除
//   - GET    /api/v1/plots/:id/assignments - 配置履歴取得
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

// CreatePlotRequest は区画作成リクエストの構造体です。
//
// フィールド:
//   - Name: 区画名（必須、最大100文字）
//   - Width: 幅（メートル、必須、0より大きい）
//   - Height: 高さ（メートル、必須、0より大きい）
//   - SoilType: 土壌タイプ（任意: clay/sandy/loamy/peaty）
//   - Sunlight: 日当たり（任意: full_sun/partial_shade/shade）
//   - PositionX: グリッドX座標（任意）
//   - PositionY: グリッドY座標（任意）
//   - Notes: メモ（任意、最大1000文字）
type CreatePlotRequest struct {
	Name      string  `json:"name" validate:"required,max=100"`
	Width     float64 `json:"width" validate:"required,gt=0"`
	Height    float64 `json:"height" validate:"required,gt=0"`
	SoilType  string  `json:"soil_type" validate:"omitempty,oneof=clay sandy loamy peaty"`
	Sunlight  string  `json:"sunlight" validate:"omitempty,oneof=full_sun partial_shade shade"`
	PositionX *int    `json:"position_x"`
	PositionY *int    `json:"position_y"`
	Notes     string  `json:"notes" validate:"max=1000"`
}

// UpdatePlotRequest は区画更新リクエストの構造体です。
// すべてのフィールドは任意で、指定されたフィールドのみ更新されます。
type UpdatePlotRequest struct {
	Name      string  `json:"name" validate:"max=100"`
	Width     float64 `json:"width" validate:"omitempty,gt=0"`
	Height    float64 `json:"height" validate:"omitempty,gt=0"`
	SoilType  string  `json:"soil_type" validate:"omitempty,oneof=clay sandy loamy peaty"`
	Sunlight  string  `json:"sunlight" validate:"omitempty,oneof=full_sun partial_shade shade"`
	PositionX *int    `json:"position_x"`
	PositionY *int    `json:"position_y"`
	Notes     string  `json:"notes" validate:"max=1000"`
}

// AssignCropRequest は作物配置リクエストの構造体です。
//
// フィールド:
//   - CropID: 配置する作物ID（必須）
//   - AssignedDate: 配置日（任意、デフォルトは現在日時）
type AssignCropRequest struct {
	CropID       uint      `json:"crop_id" validate:"required"`
	AssignedDate time.Time `json:"assigned_date"`
}

// =============================================================================
// Plot ハンドラメソッド
// =============================================================================

// GetPlots はユーザーの全区画を取得します。
//
// クエリパラメータ:
//   - status: フィルタするステータス（available/occupied）
//
// レスポンス:
//   - 200: 区画の配列
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) GetPlots(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// statusクエリパラメータでフィルタリング
	status := c.QueryParam("status")
	var plots []model.Plot
	var err error

	if status != "" {
		// ステータスでフィルタ
		plots, err = h.service.GetUserPlotsByStatus(ctx, userID, status)
	} else {
		// 全区画取得
		plots, err = h.service.GetUserPlots(ctx, userID)
	}

	if err != nil {
		return apperrors.NewInternalError("Failed to fetch plots")
	}

	return c.JSON(http.StatusOK, plots)
}

// GetPlot は特定の区画を取得します。
//
// パスパラメータ:
//   - id: 区画ID
//
// レスポンス:
//   - 200: 区画オブジェクト
//   - 400: 無効なID形式
//   - 404: 区画が見つからない
func (h *Handler) GetPlot(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid plot ID")
	}

	// 区画を取得
	plot, err := h.service.GetPlotByID(ctx, uint(id))
	if err != nil {
		return apperrors.NewNotFoundError("Plot")
	}

	return c.JSON(http.StatusOK, plot)
}

// CreatePlot は新しい区画を作成します。
//
// リクエストボディ:
//   - name: 区画名（必須）
//   - width: 幅（必須）
//   - height: 高さ（必須）
//   - soil_type: 土壌タイプ（任意）
//   - sunlight: 日当たり（任意）
//   - position_x: X座標（任意）
//   - position_y: Y座標（任意）
//   - notes: メモ（任意）
//
// レスポンス:
//   - 201: 作成された区画
//   - 400: バリデーションエラー
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) CreatePlot(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// リクエストボディをバインド&バリデーション
	var req CreatePlotRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 区画モデルを作成
	plot := &model.Plot{
		UserID:    userID,
		Name:      req.Name,
		Width:     req.Width,
		Height:    req.Height,
		SoilType:  req.SoilType,
		Sunlight:  req.Sunlight,
		Status:    "available", // 新規区画は常に available
		PositionX: req.PositionX,
		PositionY: req.PositionY,
		Notes:     req.Notes,
	}

	// DBに保存
	if err := h.service.CreatePlot(ctx, plot); err != nil {
		return apperrors.NewInternalError("Failed to create plot")
	}

	return c.JSON(http.StatusCreated, plot)
}

// UpdatePlot は既存の区画を更新します。
//
// パスパラメータ:
//   - id: 区画ID
//
// リクエストボディ: 更新するフィールド（任意）
//
// レスポンス:
//   - 200: 更新された区画
//   - 400: バリデーションエラー
//   - 404: 区画が見つからない
//   - 500: 内部エラー
func (h *Handler) UpdatePlot(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid plot ID")
	}

	// リクエストボディをバインド&バリデーション
	var req UpdatePlotRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 既存の区画を取得
	plot, err := h.service.GetPlotByID(ctx, uint(id))
	if err != nil {
		return apperrors.NewNotFoundError("Plot")
	}

	// リクエストで指定されたフィールドのみ更新
	if req.Name != "" {
		plot.Name = req.Name
	}
	if req.Width > 0 {
		plot.Width = req.Width
	}
	if req.Height > 0 {
		plot.Height = req.Height
	}
	if req.SoilType != "" {
		plot.SoilType = req.SoilType
	}
	if req.Sunlight != "" {
		plot.Sunlight = req.Sunlight
	}
	if req.PositionX != nil {
		plot.PositionX = req.PositionX
	}
	if req.PositionY != nil {
		plot.PositionY = req.PositionY
	}
	if req.Notes != "" {
		plot.Notes = req.Notes
	}

	// DBを更新
	if err := h.service.UpdatePlot(ctx, plot); err != nil {
		return apperrors.NewInternalError("Failed to update plot")
	}

	return c.JSON(http.StatusOK, plot)
}

// DeletePlot は区画を削除します（論理削除）。
// 関連する配置履歴も削除されます。
//
// パスパラメータ:
//   - id: 区画ID
//
// レスポンス:
//   - 204: 削除成功（コンテンツなし）
//   - 400: 無効なID形式
//   - 500: 内部エラー
func (h *Handler) DeletePlot(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid plot ID")
	}

	// 区画を削除（関連データも含む）
	if err := h.service.DeletePlot(ctx, uint(id)); err != nil {
		return apperrors.NewInternalError("Failed to delete plot")
	}

	return c.NoContent(http.StatusNoContent)
}

// =============================================================================
// PlotAssignment ハンドラメソッド
// =============================================================================

// AssignCrop は作物を区画に配置します。
// 既存の配置がある場合は自動的に解除されます。
//
// パスパラメータ:
//   - id: 区画ID
//
// リクエストボディ:
//   - crop_id: 配置する作物ID（必須）
//   - assigned_date: 配置日（任意、デフォルトは現在日時）
//
// レスポンス:
//   - 201: 作成された配置
//   - 400: バリデーションエラー
//   - 500: 内部エラー
func (h *Handler) AssignCrop(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	plotID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid plot ID")
	}

	// リクエストボディをバインド&バリデーション
	var req AssignCropRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 配置日が指定されていない場合は現在日時を使用
	assignedDate := req.AssignedDate
	if assignedDate.IsZero() {
		assignedDate = time.Now()
	}

	// 作物を区画に配置
	assignment, err := h.service.AssignCropToPlot(ctx, uint(plotID), req.CropID, assignedDate)
	if err != nil {
		return apperrors.NewInternalError("Failed to assign crop to plot")
	}

	return c.JSON(http.StatusCreated, assignment)
}

// UnassignCrop は区画から作物の配置を解除します。
//
// パスパラメータ:
//   - id: 区画ID
//
// レスポンス:
//   - 204: 解除成功（コンテンツなし）
//   - 400: 無効なID形式
//   - 404: アクティブな配置がない
//   - 500: 内部エラー
func (h *Handler) UnassignCrop(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	plotID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid plot ID")
	}

	// 配置を解除
	if err := h.service.UnassignCropFromPlot(ctx, uint(plotID)); err != nil {
		return apperrors.NewNotFoundError("Active assignment")
	}

	return c.NoContent(http.StatusNoContent)
}

// GetPlotAssignments は区画の全配置履歴を取得します。
//
// パスパラメータ:
//   - id: 区画ID
//
// レスポンス:
//   - 200: 配置履歴の配列
//   - 400: 無効なID形式
//   - 500: 内部エラー
func (h *Handler) GetPlotAssignments(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	plotID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid plot ID")
	}

	// 配置履歴を取得
	assignments, err := h.service.GetPlotAssignments(ctx, uint(plotID))
	if err != nil {
		return apperrors.NewInternalError("Failed to fetch plot assignments")
	}

	return c.JSON(http.StatusOK, assignments)
}

// GetActivePlotAssignment は区画の現在アクティブな配置を取得します。
//
// パスパラメータ:
//   - id: 区画ID
//
// レスポンス:
//   - 200: アクティブな配置
//   - 400: 無効なID形式
//   - 404: アクティブな配置がない
func (h *Handler) GetActivePlotAssignment(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	plotID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid plot ID")
	}

	// アクティブな配置を取得
	assignment, err := h.service.GetActivePlotAssignment(ctx, uint(plotID))
	if err != nil {
		return apperrors.NewNotFoundError("Active assignment")
	}

	return c.JSON(http.StatusOK, assignment)
}
