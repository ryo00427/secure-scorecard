package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/service"
)

// =============================================================================
// Scheduler Handler - スケジューラーハンドラー
// =============================================================================
// AWS EventBridge Scheduler から呼び出される定期タスク処理のエンドポイントを提供します。
// 認証不要で、EventBridge からの呼び出しを想定しています。

// SchedulerHandler はスケジューラー処理のハンドラーです。
type SchedulerHandler struct {
	service *service.Service
}

// NewSchedulerHandler は新しい SchedulerHandler を作成します。
func NewSchedulerHandler(svc *service.Service) *SchedulerHandler {
	return &SchedulerHandler{
		service: svc,
	}
}

// ProcessNotificationsRequest はスケジューラーリクエストの構造体です。
// EventBridge から送信される認証トークンを含みます。
type ProcessNotificationsRequest struct {
	// SchedulerToken は EventBridge Scheduler から送信される認証トークンです。
	// 環境変数 SCHEDULER_AUTH_TOKEN と一致する必要があります。
	SchedulerToken string `json:"scheduler_token" validate:"required"`
}

// ProcessNotificationsResponse はスケジューラー処理のレスポンスです。
type ProcessNotificationsResponse struct {
	Success            bool   `json:"success"`
	ProcessedAt        string `json:"processed_at"`
	OverdueTaskAlerts  int    `json:"overdue_task_alerts"`
	TodayTaskReminders int    `json:"today_task_reminders"`
	HarvestReminders   int    `json:"harvest_reminders"`
	TotalEvents        int    `json:"total_events"`
	Message            string `json:"message,omitempty"`
}

// ProcessScheduledNotifications は定期通知処理を実行します。
// AWS EventBridge Scheduler から毎日呼び出されます。
//
// エンドポイント: POST /api/v1/scheduler/notifications
//
// リクエストボディ:
//
//	{
//	  "scheduler_token": "認証トークン"
//	}
//
// レスポンス:
//
//	{
//	  "success": true,
//	  "processed_at": "2024-01-15T09:00:00Z",
//	  "overdue_task_alerts": 3,
//	  "today_task_reminders": 5,
//	  "harvest_reminders": 2,
//	  "total_events": 10,
//	  "message": "処理が正常に完了しました"
//	}
//
// 処理内容:
//   - 期限切れタスク検出（3件以上で警告通知）
//   - 当日タスクのリマインダー通知
//   - 7日以内の収穫予定リマインダー通知
//
// 注意: このエンドポイントはスケジューラー専用です。
// 認証トークンによる簡易認証を使用します。
func (h *SchedulerHandler) ProcessScheduledNotifications(c echo.Context) error {
	ctx := c.Request().Context()

	// スケジューラー処理を実行
	result, err := h.service.ProcessScheduledNotifications(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ProcessNotificationsResponse{
			Success: false,
			Message: "処理中にエラーが発生しました: " + err.Error(),
		})
	}

	// レスポンスを返す
	response := ProcessNotificationsResponse{
		Success:            true,
		ProcessedAt:        result.ProcessedAt.Format("2006-01-02T15:04:05Z07:00"),
		OverdueTaskAlerts:  result.OverdueTaskAlerts,
		TodayTaskReminders: result.TodayTaskReminders,
		HarvestReminders:   result.HarvestReminders,
		TotalEvents:        len(result.Events),
		Message:            "処理が正常に完了しました",
	}

	// TODO: NotificationService が実装されたら、result.Events を使って
	// 実際の通知（プッシュ、メール）を送信する
	// 現時点では通知イベントを生成するだけで、実際の送信は行わない

	return c.JSON(http.StatusOK, response)
}

// GetSchedulerStatus はスケジューラーのステータスを返します。
// ヘルスチェック用のエンドポイントです。
//
// エンドポイント: GET /api/v1/scheduler/status
//
// レスポンス:
//
//	{
//	  "status": "healthy",
//	  "service": "scheduler"
//	}
func (h *SchedulerHandler) GetSchedulerStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "scheduler",
	})
}

// RegisterSchedulerRoutes はスケジューラー関連のルートを登録します。
// handler.go の RegisterRoutes から呼び出されます。
func (h *Handler) RegisterSchedulerRoutes(e *echo.Echo, schedulerToken string) {
	schedulerHandler := NewSchedulerHandler(h.service)

	// スケジューラー専用エンドポイント（認証はトークンベース）
	scheduler := e.Group("/api/v1/scheduler")

	// トークン認証ミドルウェアを適用
	scheduler.Use(schedulerAuthMiddleware(schedulerToken))

	// ルート登録
	scheduler.POST("/notifications", schedulerHandler.ProcessScheduledNotifications)
	scheduler.GET("/status", schedulerHandler.GetSchedulerStatus)
}

// schedulerAuthMiddleware はスケジューラー用の簡易認証ミドルウェアです。
// リクエストヘッダーの X-Scheduler-Token と環境変数の SCHEDULER_AUTH_TOKEN を比較します。
func schedulerAuthMiddleware(expectedToken string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// トークンが設定されていない場合は認証をスキップ（開発環境用）
			if expectedToken == "" {
				return next(c)
			}

			// リクエストヘッダーからトークンを取得
			token := c.Request().Header.Get("X-Scheduler-Token")
			if token == "" {
				// ヘッダーにない場合はリクエストボディからも確認
				var req ProcessNotificationsRequest
				if err := c.Bind(&req); err == nil && req.SchedulerToken != "" {
					token = req.SchedulerToken
				}
			}

			// トークンを検証
			if token != expectedToken {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "無効な認証トークンです",
				})
			}

			return next(c)
		}
	}
}
