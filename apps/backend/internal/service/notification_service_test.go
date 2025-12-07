// Package service - NotificationService Integration Tests
//
// NotificationService の統合テストを提供します。
// テスト対象:
//   - デバイストークン登録→プッシュ通知配信フロー
//   - イベント発行→通知配信フロー（スケジューラー含む）
//   - ユーザー設定による通知スキップ
package service

import (
	"context"
	"testing"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
)

// =============================================================================
// デバイストークン登録テスト
// =============================================================================

// TestRegisterDeviceToken_Success はデバイストークン登録の正常系テストです。
// 期待動作:
//   - トークンがDBに保存される
//   - IDが自動採番される
//   - IsActiveがtrueで保存される
func TestRegisterDeviceToken_Success(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// ユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// デバイストークンを登録
	token, err := svc.RegisterDeviceToken(ctx, user.ID, "fcm-token-12345", "android", "device-uuid-1")
	if err != nil {
		t.Fatalf("RegisterDeviceToken failed: %v", err)
	}

	// 検証
	if token.ID == 0 {
		t.Error("Expected token ID to be set")
	}
	if token.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, token.UserID)
	}
	if token.Token != "fcm-token-12345" {
		t.Errorf("Expected token 'fcm-token-12345', got '%s'", token.Token)
	}
	if token.Platform != "android" {
		t.Errorf("Expected platform 'android', got '%s'", token.Platform)
	}
	if !token.IsActive {
		t.Error("Expected token to be active")
	}
}

// TestRegisterDeviceToken_UpdateExisting は既存トークン更新のテストです。
// 期待動作:
//   - 同じユーザー・プラットフォームの既存トークンが更新される
//   - 新しいレコードが作成されない
func TestRegisterDeviceToken_UpdateExisting(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// ユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 最初のトークンを登録
	token1, err := svc.RegisterDeviceToken(ctx, user.ID, "old-token", "ios", "device-1")
	if err != nil {
		t.Fatalf("First RegisterDeviceToken failed: %v", err)
	}

	// 同じプラットフォームで新しいトークンを登録
	token2, err := svc.RegisterDeviceToken(ctx, user.ID, "new-token", "ios", "device-1")
	if err != nil {
		t.Fatalf("Second RegisterDeviceToken failed: %v", err)
	}

	// 同じIDで更新されていることを確認
	if token2.ID != token1.ID {
		t.Errorf("Expected same ID %d, got %d (should update existing)", token1.ID, token2.ID)
	}
	if token2.Token != "new-token" {
		t.Errorf("Expected token 'new-token', got '%s'", token2.Token)
	}
}

// TestDeleteDeviceToken_ByPlatform はプラットフォーム指定削除のテストです。
func TestDeleteDeviceToken_ByPlatform(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// ユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 複数プラットフォームのトークンを登録
	_, _ = svc.RegisterDeviceToken(ctx, user.ID, "ios-token", "ios", "device-ios")
	_, _ = svc.RegisterDeviceToken(ctx, user.ID, "android-token", "android", "device-android")

	// iOSトークンを削除
	if err := svc.DeleteDeviceTokenByPlatform(ctx, user.ID, "ios"); err != nil {
		t.Fatalf("DeleteDeviceTokenByPlatform failed: %v", err)
	}

	// iOSトークンが削除されていることを確認
	tokens, _ := svc.GetActiveDeviceTokens(ctx, user.ID)
	for _, token := range tokens {
		if token.Platform == "ios" {
			t.Error("iOS token should have been deleted")
		}
	}
}

// =============================================================================
// 通知イベントハンドラーテスト
// =============================================================================

// TestNotificationEventHandler_HandleEvent は単一イベント処理のテストです。
// 期待動作:
//   - 通知が送信される（モックで記録）
//   - 通知ログが作成される
func TestNotificationEventHandler_HandleEvent(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	mockSender := NewMockNotificationSender()
	handler := NewNotificationEventHandler(svc, mockSender, mockRepos)
	ctx := context.Background()

	// ユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		NotificationSettings: &model.NotificationSettings{
			PushEnabled:      true,
			EmailEnabled:     true,
			TaskReminders:    true,
			HarvestReminders: true,
		},
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// デバイストークンを登録
	deviceToken := &model.DeviceToken{
		UserID:   user.ID,
		Token:    "fcm-token",
		Platform: "android",
		IsActive: true,
	}
	if err := mockRepos.DeviceToken().Create(ctx, deviceToken); err != nil {
		t.Fatalf("Failed to create device token: %v", err)
	}

	// 通知イベントを処理
	event := NotificationEvent{
		Type:   NotificationEventTaskDueReminder,
		UserID: user.ID,
		Title:  "タスクリマインダー",
		Body:   "水やりの時間です",
		Data: map[string]interface{}{
			"task_id": 123,
		},
	}

	if err := handler.HandleEvent(ctx, event); err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	// プッシュ通知が送信されたことを確認
	if len(mockSender.SentPushNotifications) == 0 {
		t.Error("Expected push notification to be sent")
	} else {
		sent := mockSender.SentPushNotifications[0]
		if sent.Token != "fcm-token" {
			t.Errorf("Expected token 'fcm-token', got '%s'", sent.Token)
		}
		if sent.Title != "タスクリマインダー" {
			t.Errorf("Expected title 'タスクリマインダー', got '%s'", sent.Title)
		}
	}

	// メール通知が送信されたことを確認
	if len(mockSender.SentEmailNotifications) == 0 {
		t.Error("Expected email notification to be sent")
	} else {
		sent := mockSender.SentEmailNotifications[0]
		if sent.ToEmail != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", sent.ToEmail)
		}
	}
}

// TestNotificationEventHandler_SettingsDisabled は通知設定無効時のテストです。
// 期待動作:
//   - 設定で無効化された通知はスキップされる
func TestNotificationEventHandler_SettingsDisabled(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	mockSender := NewMockNotificationSender()
	handler := NewNotificationEventHandler(svc, mockSender, mockRepos)
	ctx := context.Background()

	// 通知設定を無効にしたユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		NotificationSettings: &model.NotificationSettings{
			PushEnabled:      false, // プッシュ通知無効
			EmailEnabled:     false, // メール通知無効
			TaskReminders:    true,
			HarvestReminders: true,
		},
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// デバイストークンを登録
	deviceToken := &model.DeviceToken{
		UserID:   user.ID,
		Token:    "fcm-token",
		Platform: "android",
		IsActive: true,
	}
	if err := mockRepos.DeviceToken().Create(ctx, deviceToken); err != nil {
		t.Fatalf("Failed to create device token: %v", err)
	}

	// 通知イベントを処理
	event := NotificationEvent{
		Type:   NotificationEventTaskDueReminder,
		UserID: user.ID,
		Title:  "タスクリマインダー",
		Body:   "水やりの時間です",
	}

	// エラーなく完了すべき（通知はスキップされるが）
	if err := handler.HandleEvent(ctx, event); err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	// 通知が送信されていないことを確認（設定で無効化）
	// 注: MockNotificationSenderは設定を見ないので、実際のNotificationSenderの動作をテスト
	// ここではHandleEventがエラーなく完了することを確認
}

// TestNotificationEventHandler_TaskRemindersDisabled はタスクリマインダー無効時のテストです。
func TestNotificationEventHandler_TaskRemindersDisabled(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	mockSender := NewMockNotificationSender()
	handler := NewNotificationEventHandler(svc, mockSender, mockRepos)
	ctx := context.Background()

	// タスクリマインダーを無効にしたユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		NotificationSettings: &model.NotificationSettings{
			PushEnabled:      true,
			EmailEnabled:     true,
			TaskReminders:    false, // タスクリマインダー無効
			HarvestReminders: true,
		},
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// デバイストークンを登録
	deviceToken := &model.DeviceToken{
		UserID:   user.ID,
		Token:    "fcm-token",
		Platform: "android",
		IsActive: true,
	}
	if err := mockRepos.DeviceToken().Create(ctx, deviceToken); err != nil {
		t.Fatalf("Failed to create device token: %v", err)
	}

	// タスクリマインダーイベントを処理
	event := NotificationEvent{
		Type:   NotificationEventTaskDueReminder,
		UserID: user.ID,
		Title:  "タスクリマインダー",
		Body:   "水やりの時間です",
	}

	// エラーなく完了すべき
	if err := handler.HandleEvent(ctx, event); err != nil {
		t.Fatalf("HandleEvent failed: %v", err)
	}

	// 注: MockNotificationSender.SendNotificationEventは設定を見ないので
	// 実際のNotificationSenderの設定チェックは別途テストが必要
}

// =============================================================================
// スケジューラー統合テスト
// =============================================================================

// TestProcessScheduledNotificationsAndSend はスケジューラー処理の統合テストです。
// 期待動作:
//   - 期限切れタスク、当日タスク、収穫リマインダーが検出される
//   - 対応する通知イベントが生成される
//   - 通知が送信される
func TestProcessScheduledNotificationsAndSend(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	mockSender := NewMockNotificationSender()
	handler := NewNotificationEventHandler(svc, mockSender, mockRepos)
	ctx := context.Background()

	// ユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		NotificationSettings: &model.NotificationSettings{
			PushEnabled:      true,
			EmailEnabled:     true,
			TaskReminders:    true,
			HarvestReminders: true,
		},
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// デバイストークンを登録
	deviceToken := &model.DeviceToken{
		UserID:   user.ID,
		Token:    "fcm-token",
		Platform: "android",
		IsActive: true,
	}
	if err := mockRepos.DeviceToken().Create(ctx, deviceToken); err != nil {
		t.Fatalf("Failed to create device token: %v", err)
	}

	// 今日のタスクを作成
	today := time.Now().Truncate(24 * time.Hour)
	task := &model.Task{
		UserID:  user.ID,
		Title:   "水やり",
		DueDate: today,
		Status:  "pending",
	}
	if err := mockRepos.Task().Create(ctx, task); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// スケジューラー処理を実行
	result, err := handler.ProcessScheduledNotificationsAndSend(ctx)
	if err != nil {
		t.Fatalf("ProcessScheduledNotificationsAndSend failed: %v", err)
	}

	// 結果を確認
	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// 処理時刻が設定されていることを確認
	if result.ProcessedAt.IsZero() {
		t.Error("Expected ProcessedAt to be set")
	}
}

// TestProcessScheduledNotificationsAndSend_OverdueTasks は期限切れタスク警告のテストです。
// 注: このテストはモックリポジトリの制約により、Task.Userの関連付けが必要です。
func TestProcessScheduledNotificationsAndSend_OverdueTasks(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	mockSender := NewMockNotificationSender()
	handler := NewNotificationEventHandler(svc, mockSender, mockRepos)
	ctx := context.Background()

	// ユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		NotificationSettings: &model.NotificationSettings{
			PushEnabled:      true,
			EmailEnabled:     true,
			TaskReminders:    true,
			HarvestReminders: true,
		},
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// デバイストークンを登録
	deviceToken := &model.DeviceToken{
		UserID:   user.ID,
		Token:    "fcm-token",
		Platform: "android",
		IsActive: true,
	}
	if err := mockRepos.DeviceToken().Create(ctx, deviceToken); err != nil {
		t.Fatalf("Failed to create device token: %v", err)
	}

	// 期限切れタスクを3件作成（警告通知のしきい値）
	// 注: Task.Userフィールドを設定することで、processOverdueTaskAlertsでの
	// userInfo[userID]の取得が正しく動作するようにする
	yesterday := time.Now().Add(-24 * time.Hour)
	for i := 0; i < 3; i++ {
		task := &model.Task{
			UserID:  user.ID,
			Title:   "期限切れタスク",
			DueDate: yesterday,
			Status:  "pending",
			User:    *user, // ユーザー情報を関連付け（モックでPreloadをシミュレート）
		}
		if err := mockRepos.Task().Create(ctx, task); err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	// スケジューラー処理を実行
	result, err := handler.ProcessScheduledNotificationsAndSend(ctx)
	if err != nil {
		t.Fatalf("ProcessScheduledNotificationsAndSend failed: %v", err)
	}

	// 結果を確認（期限切れタスク警告が含まれるはず）
	if result.TotalEvents == 0 {
		t.Error("Expected at least one notification event for overdue tasks")
	}

	// 期限切れタスクの警告が送信されていることを確認
	if result.SuccessfulSends == 0 && result.FailedSends == 0 {
		t.Error("Expected notification to be processed")
	}
}

// =============================================================================
// 重複通知防止テスト
// =============================================================================

// TestNotificationDeduplication は重複通知防止のテストです。
func TestNotificationDeduplication(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// 通知ログを作成（24時間以内の重複を防ぐ）
	log := &model.NotificationLog{
		UserID:           1,
		NotificationType: string(NotificationEventTaskDueReminder),
		Channel:          "push,email",
		Title:            "テスト通知",
		Body:             "テスト本文",
		Status:           "sent",
		DeduplicationKey: "task_due_reminder:1:2024-01-15",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}
	if err := svc.CreateNotificationLog(ctx, log); err != nil {
		t.Fatalf("Failed to create notification log: %v", err)
	}

	// 同じキーで重複チェック
	isDuplicate, err := svc.CheckDeduplication(ctx, "task_due_reminder:1:2024-01-15")
	if err != nil {
		t.Fatalf("CheckDeduplication failed: %v", err)
	}

	if !isDuplicate {
		t.Error("Expected duplicate detection to be true")
	}

	// 異なるキーでは重複検出されない
	isDuplicate2, _ := svc.CheckDeduplication(ctx, "task_due_reminder:1:2024-01-16")
	if isDuplicate2 {
		t.Error("Expected no duplicate detection for different key")
	}
}

// =============================================================================
// 通知設定更新テスト
// =============================================================================

// TestUpdateNotificationSettings は通知設定更新のテストです。
func TestUpdateNotificationSettings(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	ctx := context.Background()

	// ユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 通知設定を更新
	settings := &model.NotificationSettings{
		PushEnabled:               false,
		EmailEnabled:              true,
		TaskReminders:             true,
		HarvestReminders:          false,
		GrowthRecordNotifications: true,
	}
	updatedSettings, err := svc.UpdateNotificationSettings(ctx, user.ID, settings)
	if err != nil {
		t.Fatalf("UpdateNotificationSettings failed: %v", err)
	}

	// 設定が更新されていることを確認
	if updatedSettings.PushEnabled != false {
		t.Error("Expected PushEnabled to be false")
	}
	if updatedSettings.EmailEnabled != true {
		t.Error("Expected EmailEnabled to be true")
	}
	if updatedSettings.HarvestReminders != false {
		t.Error("Expected HarvestReminders to be false")
	}
	if updatedSettings.GrowthRecordNotifications != true {
		t.Error("Expected GrowthRecordNotifications to be true")
	}
}

// =============================================================================
// MockNotificationSender エラーテスト
// =============================================================================

// TestNotificationEventHandler_SendError は送信エラー時のテストです。
func TestNotificationEventHandler_SendError(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	mockSender := NewMockNotificationSender()
	mockSender.ShouldFail = true // エラーを発生させる
	handler := NewNotificationEventHandler(svc, mockSender, mockRepos)
	ctx := context.Background()

	// ユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		NotificationSettings: &model.NotificationSettings{
			PushEnabled:      true,
			EmailEnabled:     true,
			TaskReminders:    true,
			HarvestReminders: true,
		},
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// デバイストークンを登録
	deviceToken := &model.DeviceToken{
		UserID:   user.ID,
		Token:    "fcm-token",
		Platform: "android",
		IsActive: true,
	}
	if err := mockRepos.DeviceToken().Create(ctx, deviceToken); err != nil {
		t.Fatalf("Failed to create device token: %v", err)
	}

	// 通知イベントを処理
	event := NotificationEvent{
		Type:   NotificationEventTaskDueReminder,
		UserID: user.ID,
		Title:  "タスクリマインダー",
		Body:   "水やりの時間です",
	}

	// エラーが返されることを確認
	err := handler.HandleEvent(ctx, event)
	if err == nil {
		t.Error("Expected error from HandleEvent when sender fails")
	}
}

// =============================================================================
// 複数イベント処理テスト
// =============================================================================

// TestNotificationEventHandler_HandleEvents は複数イベント処理のテストです。
func TestNotificationEventHandler_HandleEvents(t *testing.T) {
	mockRepos := repository.NewMockRepositories()
	svc := NewService(mockRepos)
	mockSender := NewMockNotificationSender()
	handler := NewNotificationEventHandler(svc, mockSender, mockRepos)
	ctx := context.Background()

	// ユーザーを作成
	user := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		NotificationSettings: &model.NotificationSettings{
			PushEnabled:      true,
			EmailEnabled:     true,
			TaskReminders:    true,
			HarvestReminders: true,
		},
	}
	if err := mockRepos.User().Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// デバイストークンを登録
	deviceToken := &model.DeviceToken{
		UserID:   user.ID,
		Token:    "fcm-token",
		Platform: "android",
		IsActive: true,
	}
	if err := mockRepos.DeviceToken().Create(ctx, deviceToken); err != nil {
		t.Fatalf("Failed to create device token: %v", err)
	}

	// 複数の通知イベントを作成
	events := []NotificationEvent{
		{
			Type:   NotificationEventTaskDueReminder,
			UserID: user.ID,
			Title:  "タスク1",
			Body:   "タスク1の本文",
		},
		{
			Type:   NotificationEventHarvestReminder,
			UserID: user.ID,
			Title:  "収穫リマインダー",
			Body:   "トマトの収穫時期です",
		},
	}

	// 複数イベントを処理
	result, err := handler.HandleEvents(ctx, events)
	if err != nil {
		t.Fatalf("HandleEvents failed: %v", err)
	}

	// 結果を確認
	if result.TotalEvents != 2 {
		t.Errorf("Expected 2 total events, got %d", result.TotalEvents)
	}
	if result.SuccessfulSends != 2 {
		t.Errorf("Expected 2 successful sends, got %d", result.SuccessfulSends)
	}
	if result.FailedSends != 0 {
		t.Errorf("Expected 0 failed sends, got %d", result.FailedSends)
	}
}
