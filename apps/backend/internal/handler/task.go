// Package handler - Task Handler
//
// タスク管理のHTTPハンドラを提供します。
// エンドポイント:
//   - GET    /api/v1/tasks           - ユーザーの全タスク取得
//   - GET    /api/v1/tasks/:id       - 特定のタスク取得
//   - POST   /api/v1/tasks           - 新規タスク作成
//   - PUT    /api/v1/tasks/:id       - タスク更新
//   - DELETE /api/v1/tasks/:id       - タスク削除
//   - POST   /api/v1/tasks/:id/complete - タスク完了
//   - GET    /api/v1/tasks/today     - 今日のタスク取得
//   - GET    /api/v1/tasks/overdue   - 期限切れタスク取得
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

// CreateTaskRequest はタスク作成リクエストの構造体です。
//
// フィールド:
//   - Title: タスクのタイトル（必須、最大200文字）
//   - Description: タスクの詳細説明（任意、最大1000文字）
//   - DueDate: 期限日（必須、RFC3339形式）
//   - Priority: 優先度（low/medium/high、デフォルト: medium）
//   - PlantID: 関連する植物のID（任意）
type CreateTaskRequest struct {
	Title       string    `json:"title" validate:"required,max=200"`
	Description string    `json:"description" validate:"max=1000"`
	DueDate     time.Time `json:"due_date" validate:"required"`
	Priority    string    `json:"priority" validate:"omitempty,oneof=low medium high"`
	PlantID     *uint     `json:"plant_id"`
}

// UpdateTaskRequest はタスク更新リクエストの構造体です。
// すべてのフィールドは任意で、指定されたフィールドのみ更新されます。
type UpdateTaskRequest struct {
	Title       string    `json:"title" validate:"max=200"`
	Description string    `json:"description" validate:"max=1000"`
	DueDate     time.Time `json:"due_date"`
	Priority    string    `json:"priority" validate:"omitempty,oneof=low medium high"`
	Status      string    `json:"status" validate:"omitempty,oneof=pending completed cancelled"`
	PlantID     *uint     `json:"plant_id"`
}

// =============================================================================
// ハンドラメソッド
// =============================================================================

// GetTasks はユーザーの全タスクを取得します。
//
// クエリパラメータ:
//   - status: フィルタするステータス（pending/completed/cancelled）
//
// レスポンス:
//   - 200: タスクの配列（期限日順）
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) GetTasks(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// statusクエリパラメータでフィルタリング
	status := c.QueryParam("status")
	var tasks []model.Task
	var err error

	if status != "" {
		// ステータスでフィルタ
		tasks, err = h.service.GetUserTasksByStatus(ctx, userID, status)
	} else {
		// 全タスク取得
		tasks, err = h.service.GetUserTasks(ctx, userID)
	}

	if err != nil {
		return apperrors.NewInternalError("Failed to fetch tasks")
	}

	return c.JSON(http.StatusOK, tasks)
}

// GetTask は特定のタスクを取得します。
//
// パスパラメータ:
//   - id: タスクID
//
// レスポンス:
//   - 200: タスクオブジェクト
//   - 400: 無効なID形式
//   - 404: タスクが見つからない
func (h *Handler) GetTask(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid task ID")
	}

	// タスクを取得
	task, err := h.service.GetTaskByID(ctx, uint(id))
	if err != nil {
		return apperrors.NewNotFoundError("Task")
	}

	return c.JSON(http.StatusOK, task)
}

// CreateTask は新しいタスクを作成します。
//
// リクエストボディ:
//   - title: タスクタイトル（必須）
//   - description: 説明（任意）
//   - due_date: 期限日（必須）
//   - priority: 優先度（任意、デフォルト: medium）
//   - plant_id: 関連植物ID（任意）
//
// レスポンス:
//   - 201: 作成されたタスク
//   - 400: バリデーションエラー
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) CreateTask(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// リクエストボディをバインド&バリデーション
	var req CreateTaskRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 優先度のデフォルト値を設定
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	// タスクモデルを作成
	task := &model.Task{
		UserID:      userID,
		PlantID:     req.PlantID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Priority:    priority,
		Status:      "pending", // 新規タスクは常に pending
	}

	// DBに保存
	if err := h.service.CreateTask(ctx, task); err != nil {
		return apperrors.NewInternalError("Failed to create task")
	}

	return c.JSON(http.StatusCreated, task)
}

// UpdateTask は既存のタスクを更新します。
//
// パスパラメータ:
//   - id: タスクID
//
// リクエストボディ: 更新するフィールド（任意）
//
// レスポンス:
//   - 200: 更新されたタスク
//   - 400: バリデーションエラー
//   - 404: タスクが見つからない
//   - 500: 内部エラー
func (h *Handler) UpdateTask(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid task ID")
	}

	// リクエストボディをバインド&バリデーション
	var req UpdateTaskRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// 既存のタスクを取得
	task, err := h.service.GetTaskByID(ctx, uint(id))
	if err != nil {
		return apperrors.NewNotFoundError("Task")
	}

	// リクエストで指定されたフィールドのみ更新
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if !req.DueDate.IsZero() {
		task.DueDate = req.DueDate
	}
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.Status != "" {
		task.Status = req.Status
		// completedに変更された場合、CompletedAtを設定
		if req.Status == "completed" && task.CompletedAt == nil {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if req.PlantID != nil {
		task.PlantID = req.PlantID
	}

	// DBを更新
	if err := h.service.UpdateTask(ctx, task); err != nil {
		return apperrors.NewInternalError("Failed to update task")
	}

	return c.JSON(http.StatusOK, task)
}

// DeleteTask はタスクを削除します（論理削除）。
//
// パスパラメータ:
//   - id: タスクID
//
// レスポンス:
//   - 204: 削除成功（コンテンツなし）
//   - 400: 無効なID形式
//   - 500: 内部エラー
func (h *Handler) DeleteTask(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid task ID")
	}

	// タスクを削除
	if err := h.service.DeleteTask(ctx, uint(id)); err != nil {
		return apperrors.NewInternalError("Failed to delete task")
	}

	return c.NoContent(http.StatusNoContent)
}

// CompleteTask はタスクを完了としてマークします。
//
// パスパラメータ:
//   - id: タスクID
//
// レスポンス:
//   - 200: 完了したタスク
//   - 400: 無効なID形式
//   - 404: タスクが見つからない
//   - 500: 内部エラー
func (h *Handler) CompleteTask(c echo.Context) error {
	ctx := c.Request().Context()

	// パスパラメータからIDを取得
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return apperrors.NewBadRequestError("Invalid task ID")
	}

	// タスクを完了
	if err := h.service.CompleteTask(ctx, uint(id)); err != nil {
		return apperrors.NewNotFoundError("Task")
	}

	// 更新後のタスクを取得して返す
	task, err := h.service.GetTaskByID(ctx, uint(id))
	if err != nil {
		return apperrors.NewInternalError("Failed to fetch completed task")
	}

	return c.JSON(http.StatusOK, task)
}

// GetTodayTasks は今日が期限のタスクを取得します。
// ダッシュボード用のエンドポイントです。
//
// レスポンス:
//   - 200: 今日のタスクの配列（優先度順）
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) GetTodayTasks(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// 今日のタスクを取得
	tasks, err := h.service.GetTodayTasks(ctx, userID)
	if err != nil {
		return apperrors.NewInternalError("Failed to fetch today's tasks")
	}

	return c.JSON(http.StatusOK, tasks)
}

// GetOverdueTasks は期限切れのタスクを取得します。
// ダッシュボード用のエンドポイントです。
//
// レスポンス:
//   - 200: 期限切れタスクの配列
//   - 401: 認証エラー
//   - 500: 内部エラー
func (h *Handler) GetOverdueTasks(c echo.Context) error {
	ctx := c.Request().Context()

	// 認証済みユーザーIDを取得
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		return apperrors.NewAuthenticationError("Not authenticated")
	}

	// 期限切れタスクを取得
	tasks, err := h.service.GetOverdueTasks(ctx, userID)
	if err != nil {
		return apperrors.NewInternalError("Failed to fetch overdue tasks")
	}

	return c.JSON(http.StatusOK, tasks)
}
