package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
)

// =============================================================================
// Notification Processor - 通知処理サービス
// =============================================================================
// スケジューラーからのイベントを受け取り、実際の通知配信を行います。
// ユーザーの通知設定を確認し、重複防止処理も行います。

// Processor は通知処理サービスです
type Processor struct {
	repos  repository.Repositories
	sender *Sender
}

// NewProcessor は新しい通知処理サービスを作成します
func NewProcessor(repos repository.Repositories, sender *Sender) *Processor {
	return &Processor{
		repos:  repos,
		sender: sender,
	}
}

// ProcessingResult は処理結果を表します
type ProcessingResult struct {
	TotalEvents     int            `json:"total_events"`
	ProcessedCount  int            `json:"processed_count"`
	SkippedCount    int            `json:"skipped_count"`     // 設定によりスキップされた件数
	DeduplicatedCount int          `json:"deduplicated_count"` // 重複防止でスキップされた件数
	SuccessCount    int            `json:"success_count"`
	FailedCount     int            `json:"failed_count"`
	Errors          []string       `json:"errors,omitempty"`
	ProcessedAt     time.Time      `json:"processed_at"`
}

// NotificationEvent は通知イベントを表します（service層から受け取る）
type NotificationEvent struct {
	Type      string                 `json:"type"`
	UserID    uint                   `json:"user_id"`
	UserEmail string                 `json:"user_email"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// =============================================================================
// Event Processing Methods - イベント処理メソッド
// =============================================================================

// ProcessEvents は通知イベントを処理して配信します
// ユーザー設定の確認、重複防止チェック、実際の配信を行います
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - events: 処理する通知イベント
//
// 戻り値:
//   - *ProcessingResult: 処理結果
func (p *Processor) ProcessEvents(ctx context.Context, events []NotificationEvent) *ProcessingResult {
	result := &ProcessingResult{
		TotalEvents: len(events),
		ProcessedAt: time.Now(),
		Errors:      make([]string, 0),
	}

	for _, event := range events {
		// 1. ユーザー情報と通知設定を取得
		user, err := p.repos.User().GetByID(ctx, event.UserID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("user %d not found: %v", event.UserID, err))
			result.FailedCount++
			continue
		}

		// 2. 通知設定をチェック
		if !p.shouldSendNotification(user, event.Type) {
			result.SkippedCount++
			continue
		}

		// 3. 重複防止チェック
		deduplicationKey := p.generateDeduplicationKey(event)
		isDuplicate, err := p.checkDeduplication(ctx, deduplicationKey)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("deduplication check failed: %v", err))
		}
		if isDuplicate {
			result.DeduplicatedCount++
			continue
		}

		result.ProcessedCount++

		// 4. 通知を送信
		success := p.sendNotification(ctx, user, event)

		// 5. 通知ログを記録
		p.recordNotificationLog(ctx, user.ID, event, deduplicationKey, success)

		if success {
			result.SuccessCount++
		} else {
			result.FailedCount++
		}
	}

	return result
}

// shouldSendNotification はユーザーの通知設定に基づいて通知を送るべきか判定します
func (p *Processor) shouldSendNotification(user *model.User, eventType string) bool {
	settings := user.NotificationSettings
	if settings == nil {
		// デフォルトは通知有効
		return true
	}

	// プッシュ通知もメール通知も無効な場合はスキップ
	if !settings.PushEnabled && !settings.EmailEnabled {
		return false
	}

	// イベントタイプに応じた設定チェック
	switch eventType {
	case "task_due_reminder", "task_overdue_alert":
		return settings.TaskReminders
	case "harvest_reminder":
		return settings.HarvestReminders
	case "growth_record_added":
		return settings.GrowthRecordNotifications
	default:
		return true
	}
}

// generateDeduplicationKey は重複防止用のキーを生成します
// フォーマット: {eventType}:{userID}:{date}
func (p *Processor) generateDeduplicationKey(event NotificationEvent) string {
	date := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s:%d:%s", event.Type, event.UserID, date)
}

// checkDeduplication は重複通知かどうかをチェックします
func (p *Processor) checkDeduplication(ctx context.Context, key string) (bool, error) {
	log, err := p.repos.NotificationLog().GetByDeduplicationKey(ctx, key)
	if err != nil {
		// レコードが見つからない場合は重複ではない
		return false, nil
	}
	return log != nil, nil
}

// sendNotification は実際の通知を送信します
// プッシュ通知とメール通知の両方を試みます
func (p *Processor) sendNotification(ctx context.Context, user *model.User, event NotificationEvent) bool {
	settings := user.NotificationSettings
	var pushSuccess, emailSuccess bool

	// プッシュ通知を送信
	if settings == nil || settings.PushEnabled {
		pushSuccess = p.sendPushNotification(ctx, user.ID, event)
	}

	// メール通知を送信
	if settings == nil || settings.EmailEnabled {
		emailSuccess = p.sendEmailNotification(ctx, user.Email, event)
	}

	return pushSuccess || emailSuccess
}

// sendPushNotification はプッシュ通知を送信します
func (p *Processor) sendPushNotification(ctx context.Context, userID uint, event NotificationEvent) bool {
	if p.sender == nil {
		return false
	}

	// ユーザーのアクティブなデバイストークンを取得
	tokens, err := p.repos.DeviceToken().GetActiveByUserID(ctx, userID)
	if err != nil || len(tokens) == 0 {
		return false
	}

	message := PushMessage{
		Title:    event.Title,
		Body:     event.Body,
		Data:     event.Data,
		Priority: "high",
	}

	// 各デバイスに送信
	var anySuccess bool
	for _, token := range tokens {
		// SNS Endpoint ARNが必要（実際の実装ではトークンからエンドポイントを作成/取得）
		endpointARN := token.Token // 簡易実装（本来はSNS Endpoint ARNを別途管理）
		result := p.sender.SendPush(ctx, endpointARN, message)
		if result.Success {
			anySuccess = true
		} else {
			// 無効なトークンの場合は無効化
			if isInvalidTokenError(result.Error) {
				_ = p.repos.DeviceToken().DeactivateToken(ctx, token.ID)
			}
		}
	}

	return anySuccess
}

// sendEmailNotification はメール通知を送信します
func (p *Processor) sendEmailNotification(ctx context.Context, email string, event NotificationEvent) bool {
	if p.sender == nil || email == "" {
		return false
	}

	// イベントタイプに応じたメールテンプレートを使用
	var emailMsg EmailMessage
	switch event.Type {
	case "task_due_reminder":
		taskCount := 1
		if data, ok := event.Data["task_count"]; ok {
			if count, ok := data.(int); ok {
				taskCount = count
			}
		}
		emailMsg = TaskReminderEmailTemplate(event.Title, time.Now().Format("2006-01-02"), taskCount)
	case "task_overdue_alert":
		overdueCount := 3
		if data, ok := event.Data["overdue_count"]; ok {
			if count, ok := data.(int); ok {
				overdueCount = count
			}
		}
		emailMsg = OverdueAlertEmailTemplate(overdueCount)
	case "harvest_reminder":
		daysUntil := 7
		if data, ok := event.Data["days_until"]; ok {
			if days, ok := data.(int); ok {
				daysUntil = days
			}
		}
		cropName := event.Title
		if data, ok := event.Data["crop_name"]; ok {
			if name, ok := data.(string); ok {
				cropName = name
			}
		}
		emailMsg = HarvestReminderEmailTemplate(cropName, daysUntil)
	default:
		emailMsg = EmailMessage{
			Subject:  event.Title,
			BodyText: event.Body,
		}
	}

	emailMsg.To = email
	result := p.sender.SendEmail(ctx, emailMsg)
	return result.Success
}

// recordNotificationLog は通知ログを記録します
func (p *Processor) recordNotificationLog(ctx context.Context, userID uint, event NotificationEvent, deduplicationKey string, success bool) {
	status := "sent"
	if !success {
		status = "failed"
	}

	now := time.Now()
	log := &model.NotificationLog{
		UserID:           userID,
		NotificationType: event.Type,
		Channel:          "push,email",
		Title:            event.Title,
		Body:             event.Body,
		Status:           status,
		DeduplicationKey: deduplicationKey,
		ExpiresAt:        now.Add(24 * time.Hour), // TTL: 24時間
	}
	if success {
		log.SentAt = &now
	}

	_ = p.repos.NotificationLog().Create(ctx, log)
}

// isInvalidTokenError はトークンが無効かどうかを判定します
func isInvalidTokenError(errMsg string) bool {
	// SNSからの無効トークンエラーを検出
	// 実際のエラーメッセージに応じて調整
	return errMsg != "" && (
		contains(errMsg, "InvalidPlatformToken") ||
		contains(errMsg, "EndpointDisabled") ||
		contains(errMsg, "InvalidRegistration"))
}

// contains は文字列に部分文字列が含まれるかチェックします
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
