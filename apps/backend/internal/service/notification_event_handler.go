package service

import (
	"context"
	"fmt"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
)

// =============================================================================
// Notification Event Handler - 通知イベントハンドラー
// =============================================================================
// スケジューラーから生成された通知イベントを処理し、
// 実際のプッシュ通知・メール通知を送信します。

// NotificationEventHandler は通知イベント処理インターフェースです。
type NotificationEventHandler interface {
	// HandleEvent は単一の通知イベントを処理します。
	HandleEvent(ctx context.Context, event NotificationEvent) error

	// HandleEvents は複数の通知イベントを処理します。
	HandleEvents(ctx context.Context, events []NotificationEvent) (*NotificationProcessResult, error)

	// ProcessScheduledNotificationsAndSend はスケジューラー処理と通知送信を実行します。
	ProcessScheduledNotificationsAndSend(ctx context.Context) (*NotificationProcessResult, error)
}

// NotificationProcessResult は通知処理の結果を表します。
type NotificationProcessResult struct {
	ProcessedAt     time.Time `json:"processed_at"`
	TotalEvents     int       `json:"total_events"`
	SuccessfulSends int       `json:"successful_sends"`
	FailedSends     int       `json:"failed_sends"`
	SkippedSends    int       `json:"skipped_sends"` // 設定で無効化されたもの
	Errors          []string  `json:"errors,omitempty"`
}

// notificationEventHandler はNotificationEventHandlerの実装です。
type notificationEventHandler struct {
	service *Service
	sender  NotificationSender
	repos   repository.Repositories
}

// NewNotificationEventHandler は新しいNotificationEventHandlerを作成します。
//
// 引数:
//   - service: サービス層（スケジューラー処理用）
//   - sender: 通知送信インターフェース
//   - repos: リポジトリ（ユーザー・トークン取得用）
//
// 戻り値:
//   - NotificationEventHandler: イベントハンドラー
func NewNotificationEventHandler(service *Service, sender NotificationSender, repos repository.Repositories) NotificationEventHandler {
	return &notificationEventHandler{
		service: service,
		sender:  sender,
		repos:   repos,
	}
}

// HandleEvent は単一の通知イベントを処理します。
// ユーザー情報とデバイストークンを取得し、通知を送信します。
//
// 引数:
//   - ctx: コンテキスト
//   - event: 通知イベント
//
// 戻り値:
//   - error: 処理に失敗した場合のエラー
func (h *notificationEventHandler) HandleEvent(ctx context.Context, event NotificationEvent) error {
	// ユーザー情報を取得
	user, err := h.repos.User().GetByID(ctx, event.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", event.UserID, err)
	}

	// 重複チェック
	deduplicationKey := generateDeduplicationKey(event)
	isDuplicate, err := h.service.CheckDeduplication(ctx, deduplicationKey)
	if err == nil && isDuplicate {
		return nil // 重複のためスキップ
	}

	// デバイストークンを取得
	tokens, err := h.repos.DeviceToken().GetActiveByUserID(ctx, event.UserID)
	if err != nil {
		// トークンがなくてもメール通知は可能なのでエラーとしない
		tokens = []model.DeviceToken{}
	}

	// 通知を送信
	sendErr := h.sender.SendNotificationEvent(ctx, event, user, tokens)

	// 通知ログを記録
	status := "sent"
	var errorMessage string
	if sendErr != nil {
		status = "failed"
		errorMessage = sendErr.Error()
	}

	log := &model.NotificationLog{
		UserID:           event.UserID,
		NotificationType: string(event.Type),
		Channel:          "push,email",
		Title:            event.Title,
		Body:             event.Body,
		Status:           status,
		ErrorMessage:     errorMessage,
		DeduplicationKey: deduplicationKey,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}
	if status == "sent" {
		now := time.Now()
		log.SentAt = &now
	}

	if logErr := h.service.CreateNotificationLog(ctx, log); logErr != nil {
		// ログ記録失敗は警告レベルとして処理を継続
		fmt.Printf("warning: failed to create notification log: %v\n", logErr)
	}

	return sendErr
}

// HandleEvents は複数の通知イベントを処理します。
// 各イベントを順次処理し、結果を集計して返します。
//
// 引数:
//   - ctx: コンテキスト
//   - events: 通知イベントのリスト
//
// 戻り値:
//   - *NotificationProcessResult: 処理結果
//   - error: 致命的なエラーが発生した場合
func (h *notificationEventHandler) HandleEvents(ctx context.Context, events []NotificationEvent) (*NotificationProcessResult, error) {
	result := &NotificationProcessResult{
		ProcessedAt: time.Now(),
		TotalEvents: len(events),
		Errors:      make([]string, 0),
	}

	for _, event := range events {
		if err := h.HandleEvent(ctx, event); err != nil {
			result.FailedSends++
			result.Errors = append(result.Errors, fmt.Sprintf("event %s for user %d: %v", event.Type, event.UserID, err))
		} else {
			result.SuccessfulSends++
		}
	}

	return result, nil
}

// ProcessScheduledNotificationsAndSend はスケジューラー処理と通知送信を実行します。
// このメソッドはEventBridge Schedulerから定期的に呼び出されます。
//
// 処理フロー:
//  1. ProcessScheduledNotifications を呼び出してイベントを生成
//  2. 生成されたイベントを HandleEvents で処理
//
// 引数:
//   - ctx: コンテキスト
//
// 戻り値:
//   - *NotificationProcessResult: 処理結果
//   - error: 致命的なエラーが発生した場合
func (h *notificationEventHandler) ProcessScheduledNotificationsAndSend(ctx context.Context) (*NotificationProcessResult, error) {
	// 1. スケジューラー処理でイベントを生成
	schedulerResult, err := h.service.ProcessScheduledNotifications(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to process scheduled notifications: %w", err)
	}

	// 2. 生成されたイベントを処理
	result, err := h.HandleEvents(ctx, schedulerResult.Events)
	if err != nil {
		return nil, fmt.Errorf("failed to handle events: %w", err)
	}

	return result, nil
}

// generateDeduplicationKey は通知イベントの重複防止キーを生成します。
// 24時間以内に同じキーで送信された通知はスキップされます。
//
// キーのフォーマット: {event_type}:{user_id}:{date}
func generateDeduplicationKey(event NotificationEvent) string {
	today := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s:%d:%s", event.Type, event.UserID, today)
}

// =============================================================================
// Scheduler Endpoint Handler - スケジューラーエンドポイント
// =============================================================================
// EventBridge Scheduler からの HTTP リクエストを処理します。

// SchedulerHandler はスケジューラーエンドポイント用のハンドラー設定を提供します。
// このメソッドはhandler層から呼び出されます。

// ProcessDailyNotifications は日次通知処理を実行します。
// EventBridge Scheduler から毎日呼び出されることを想定しています。
//
// 処理内容:
//   - 期限切れタスクの警告通知（3件以上の場合）
//   - 当日タスクのリマインダー通知
//   - 7日以内の収穫予定リマインダー通知
//
// 引数:
//   - ctx: コンテキスト
//   - handler: 通知イベントハンドラー
//
// 戻り値:
//   - *NotificationProcessResult: 処理結果
//   - error: 処理に失敗した場合のエラー
func ProcessDailyNotifications(ctx context.Context, handler NotificationEventHandler) (*NotificationProcessResult, error) {
	return handler.ProcessScheduledNotificationsAndSend(ctx)
}
