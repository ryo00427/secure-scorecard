package repository

import (
	"context"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// =============================================================================
// DeviceTokenRepository Implementation - デバイストークンリポジトリ
// =============================================================================

// deviceTokenRepository implements DeviceTokenRepository
type deviceTokenRepository struct {
	db *gorm.DB
}

// Create は新しいデバイストークンを登録します。
// 同じユーザー・プラットフォームの既存トークンがある場合は更新します。
func (r *deviceTokenRepository) Create(ctx context.Context, token *model.DeviceToken) error {
	return GetDB(ctx, r.db).Create(token).Error
}

// GetByID はIDでデバイストークンを取得します。
func (r *deviceTokenRepository) GetByID(ctx context.Context, id uint) (*model.DeviceToken, error) {
	var token model.DeviceToken
	if err := GetDB(ctx, r.db).First(&token, id).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// GetByUserID はユーザーの全デバイストークンを取得します。
func (r *deviceTokenRepository) GetByUserID(ctx context.Context, userID uint) ([]model.DeviceToken, error) {
	var tokens []model.DeviceToken
	if err := GetDB(ctx, r.db).Where("user_id = ?", userID).Order("updated_at DESC").Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// GetByUserIDAndPlatform はユーザーとプラットフォームでトークンを取得します。
// (userID, platform) で1レコードを想定（upsert用）
func (r *deviceTokenRepository) GetByUserIDAndPlatform(ctx context.Context, userID uint, platform string) (*model.DeviceToken, error) {
	var token model.DeviceToken
	if err := GetDB(ctx, r.db).Where("user_id = ? AND platform = ?", userID, platform).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// GetByToken はトークン文字列でデバイストークンを取得します。
// トークンの重複チェックに使用します。
func (r *deviceTokenRepository) GetByToken(ctx context.Context, token string) (*model.DeviceToken, error) {
	var deviceToken model.DeviceToken
	if err := GetDB(ctx, r.db).Where("token = ?", token).First(&deviceToken).Error; err != nil {
		return nil, err
	}
	return &deviceToken, nil
}

// GetActiveByUserID はユーザーのアクティブなトークンを取得します。
// 通知送信時に使用します。
func (r *deviceTokenRepository) GetActiveByUserID(ctx context.Context, userID uint) ([]model.DeviceToken, error) {
	var tokens []model.DeviceToken
	if err := GetDB(ctx, r.db).Where("user_id = ? AND is_active = ?", userID, true).Order("updated_at DESC").Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// Update はデバイストークンを更新します。
func (r *deviceTokenRepository) Update(ctx context.Context, token *model.DeviceToken) error {
	return GetDB(ctx, r.db).Save(token).Error
}

// Delete はデバイストークンを削除します。
func (r *deviceTokenRepository) Delete(ctx context.Context, id uint) error {
	return GetDB(ctx, r.db).Delete(&model.DeviceToken{}, id).Error
}

// DeleteByUserID はユーザーの全デバイストークンを削除します。
// ユーザー削除時やログアウト時に使用します。
func (r *deviceTokenRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return GetDB(ctx, r.db).Where("user_id = ?", userID).Delete(&model.DeviceToken{}).Error
}

// DeactivateToken はトークンを無効化します。
// 無効なトークンが検出された場合（SNSからのエラー等）に使用します。
func (r *deviceTokenRepository) DeactivateToken(ctx context.Context, id uint) error {
	return GetDB(ctx, r.db).Model(&model.DeviceToken{}).Where("id = ?", id).Update("is_active", false).Error
}

// =============================================================================
// NotificationLogRepository Implementation - 通知ログリポジトリ
// =============================================================================

// notificationLogRepository implements NotificationLogRepository
type notificationLogRepository struct {
	db *gorm.DB
}

// Create は新しい通知ログを作成します。
func (r *notificationLogRepository) Create(ctx context.Context, log *model.NotificationLog) error {
	return GetDB(ctx, r.db).Create(log).Error
}

// GetByID はIDで通知ログを取得します。
func (r *notificationLogRepository) GetByID(ctx context.Context, id uint) (*model.NotificationLog, error) {
	var log model.NotificationLog
	if err := GetDB(ctx, r.db).First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByDeduplicationKey は重複防止キーで通知ログを取得します。
// 24時間以内に同じキーで送信された通知があるかチェックします。
func (r *notificationLogRepository) GetByDeduplicationKey(ctx context.Context, key string) (*model.NotificationLog, error) {
	var log model.NotificationLog
	// 期限切れでない、同じ重複防止キーのログを検索
	if err := GetDB(ctx, r.db).
		Where("deduplication_key = ? AND expires_at > ?", key, time.Now()).
		First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByUserID はユーザーの通知ログを取得します。
// 最新順にソートして取得します。
func (r *notificationLogRepository) GetByUserID(ctx context.Context, userID uint, limit int) ([]model.NotificationLog, error) {
	var logs []model.NotificationLog
	query := GetDB(ctx, r.db).Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetPendingNotifications は送信待ちの通知を取得します。
// リトライ処理で使用します。
func (r *notificationLogRepository) GetPendingNotifications(ctx context.Context, limit int) ([]model.NotificationLog, error) {
	var logs []model.NotificationLog
	query := GetDB(ctx, r.db).
		Where("status = ? AND retry_count < ?", "pending", 3).
		Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// Update は通知ログを更新します。
func (r *notificationLogRepository) Update(ctx context.Context, log *model.NotificationLog) error {
	return GetDB(ctx, r.db).Save(log).Error
}

// DeleteExpired は期限切れの通知ログを削除します。
// 定期的なクリーンアップジョブで使用します。
func (r *notificationLogRepository) DeleteExpired(ctx context.Context) error {
	return GetDB(ctx, r.db).Where("expires_at < ?", time.Now()).Delete(&model.NotificationLog{}).Error
}
