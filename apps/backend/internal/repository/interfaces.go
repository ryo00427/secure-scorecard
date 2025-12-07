package repository

import (
	"context"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByFirebaseUID(ctx context.Context, uid string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
}

// GardenRepository defines the interface for garden data access
type GardenRepository interface {
	Create(ctx context.Context, garden *model.Garden) error
	GetByID(ctx context.Context, id uint) (*model.Garden, error)
	GetByUserID(ctx context.Context, userID uint) ([]model.Garden, error)
	Update(ctx context.Context, garden *model.Garden) error
	Delete(ctx context.Context, id uint) error
}

// PlantRepository defines the interface for plant data access
type PlantRepository interface {
	Create(ctx context.Context, plant *model.Plant) error
	GetByID(ctx context.Context, id uint) (*model.Plant, error)
	GetByGardenID(ctx context.Context, gardenID uint) ([]model.Plant, error)
	Update(ctx context.Context, plant *model.Plant) error
	Delete(ctx context.Context, id uint) error
	DeleteByGardenID(ctx context.Context, gardenID uint) error
}

// CareLogRepository defines the interface for care log data access
type CareLogRepository interface {
	Create(ctx context.Context, careLog *model.CareLog) error
	GetByID(ctx context.Context, id uint) (*model.CareLog, error)
	GetByPlantID(ctx context.Context, plantID uint) ([]model.CareLog, error)
	Delete(ctx context.Context, id uint) error
}

// TokenBlacklistRepository defines the interface for token blacklist data access
type TokenBlacklistRepository interface {
	Add(ctx context.Context, tokenHash string, expiresAt time.Time) error
	IsBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	DeleteExpired(ctx context.Context) error
}

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id uint) (*model.Task, error)
	GetByUserID(ctx context.Context, userID uint) ([]model.Task, error)
	GetByUserIDAndStatus(ctx context.Context, userID uint, status string) ([]model.Task, error)
	GetTodayTasks(ctx context.Context, userID uint) ([]model.Task, error)
	GetOverdueTasks(ctx context.Context, userID uint) ([]model.Task, error)
	// GetAllOverdueTasks はシステム全体の期限切れタスクを取得します（通知処理用）
	GetAllOverdueTasks(ctx context.Context) ([]model.Task, error)
	// GetAllTodayTasks はシステム全体の今日が期限のタスクを取得します（通知処理用）
	GetAllTodayTasks(ctx context.Context) ([]model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id uint) error
}

// CropRepository defines the interface for crop data access
// 作物の植え付けから収穫までのライフサイクルを管理します
type CropRepository interface {
	Create(ctx context.Context, crop *model.Crop) error
	GetByID(ctx context.Context, id uint) (*model.Crop, error)
	GetByUserID(ctx context.Context, userID uint) ([]model.Crop, error)
	GetByUserIDAndStatus(ctx context.Context, userID uint, status string) ([]model.Crop, error)
	// GetUpcomingHarvests は指定日数以内に収穫予定の作物を取得します（通知処理用）
	GetUpcomingHarvests(ctx context.Context, daysAhead int) ([]model.Crop, error)
	Update(ctx context.Context, crop *model.Crop) error
	Delete(ctx context.Context, id uint) error
}

// GrowthRecordRepository defines the interface for growth record data access
// 作物の成長記録を管理します
type GrowthRecordRepository interface {
	Create(ctx context.Context, record *model.GrowthRecord) error
	GetByID(ctx context.Context, id uint) (*model.GrowthRecord, error)
	GetByCropID(ctx context.Context, cropID uint) ([]model.GrowthRecord, error)
	Delete(ctx context.Context, id uint) error
	DeleteByCropID(ctx context.Context, cropID uint) error
}

// HarvestRepository defines the interface for harvest data access
// 収穫記録を管理します
type HarvestRepository interface {
	Create(ctx context.Context, harvest *model.Harvest) error
	GetByID(ctx context.Context, id uint) (*model.Harvest, error)
	GetByCropID(ctx context.Context, cropID uint) ([]model.Harvest, error)
	// GetByUserIDWithDateRange はユーザーの収穫記録を日付範囲でフィルタして取得します
	// Analytics用。startDate/endDateがnilの場合は制限なし
	GetByUserIDWithDateRange(ctx context.Context, userID uint, startDate, endDate *time.Time) ([]model.Harvest, error)
	Delete(ctx context.Context, id uint) error
	DeleteByCropID(ctx context.Context, cropID uint) error
}

// PlotRepository defines the interface for plot data access
// 菜園の区画を管理します（グリッドレイアウト対応）
type PlotRepository interface {
	Create(ctx context.Context, plot *model.Plot) error
	GetByID(ctx context.Context, id uint) (*model.Plot, error)
	GetByUserID(ctx context.Context, userID uint) ([]model.Plot, error)
	GetByUserIDAndStatus(ctx context.Context, userID uint, status string) ([]model.Plot, error)
	Update(ctx context.Context, plot *model.Plot) error
	Delete(ctx context.Context, id uint) error
}

// PlotAssignmentRepository defines the interface for plot assignment data access
// 区画への作物配置を管理します（履歴追跡対応）
type PlotAssignmentRepository interface {
	Create(ctx context.Context, assignment *model.PlotAssignment) error
	GetByID(ctx context.Context, id uint) (*model.PlotAssignment, error)
	GetByPlotID(ctx context.Context, plotID uint) ([]model.PlotAssignment, error)
	GetActiveByPlotID(ctx context.Context, plotID uint) (*model.PlotAssignment, error) // 現在アクティブな配置
	GetByCropID(ctx context.Context, cropID uint) ([]model.PlotAssignment, error)
	Update(ctx context.Context, assignment *model.PlotAssignment) error
	Delete(ctx context.Context, id uint) error
	DeleteByPlotID(ctx context.Context, plotID uint) error
}

// DeviceTokenRepository defines the interface for device token data access
// プッシュ通知用のデバイストークンを管理します
type DeviceTokenRepository interface {
	// Create は新しいデバイストークンを登録します
	Create(ctx context.Context, token *model.DeviceToken) error
	// GetByID はIDでデバイストークンを取得します
	GetByID(ctx context.Context, id uint) (*model.DeviceToken, error)
	// GetByUserID はユーザーの全デバイストークンを取得します
	GetByUserID(ctx context.Context, userID uint) ([]model.DeviceToken, error)
	// GetByUserIDAndPlatform はユーザーとプラットフォームでトークンを取得します
	GetByUserIDAndPlatform(ctx context.Context, userID uint, platform string) (*model.DeviceToken, error)
	// GetByToken はトークン文字列でデバイストークンを取得します
	GetByToken(ctx context.Context, token string) (*model.DeviceToken, error)
	// GetActiveByUserID はユーザーのアクティブなトークンを取得します
	GetActiveByUserID(ctx context.Context, userID uint) ([]model.DeviceToken, error)
	// Update はデバイストークンを更新します
	Update(ctx context.Context, token *model.DeviceToken) error
	// Delete はデバイストークンを削除します
	Delete(ctx context.Context, id uint) error
	// DeleteByUserID はユーザーの全デバイストークンを削除します
	DeleteByUserID(ctx context.Context, userID uint) error
	// DeactivateToken はトークンを無効化します（無効トークン検出時）
	DeactivateToken(ctx context.Context, id uint) error
}

// NotificationLogRepository defines the interface for notification log data access
// 通知の送信ログを管理します（重複防止、配信追跡）
type NotificationLogRepository interface {
	// Create は新しい通知ログを作成します
	Create(ctx context.Context, log *model.NotificationLog) error
	// GetByID はIDで通知ログを取得します
	GetByID(ctx context.Context, id uint) (*model.NotificationLog, error)
	// GetByDeduplicationKey は重複防止キーで通知ログを取得します
	GetByDeduplicationKey(ctx context.Context, key string) (*model.NotificationLog, error)
	// GetByUserID はユーザーの通知ログを取得します
	GetByUserID(ctx context.Context, userID uint, limit int) ([]model.NotificationLog, error)
	// GetPendingNotifications は送信待ちの通知を取得します（リトライ用）
	GetPendingNotifications(ctx context.Context, limit int) ([]model.NotificationLog, error)
	// Update は通知ログを更新します
	Update(ctx context.Context, log *model.NotificationLog) error
	// DeleteExpired は期限切れの通知ログを削除します
	DeleteExpired(ctx context.Context) error
}

// Repositories aggregates all repository interfaces
type Repositories interface {
	User() UserRepository
	Garden() GardenRepository
	Plant() PlantRepository
	CareLog() CareLogRepository
	TokenBlacklist() TokenBlacklistRepository
	Task() TaskRepository
	Crop() CropRepository
	GrowthRecord() GrowthRecordRepository
	Harvest() HarvestRepository
	Plot() PlotRepository
	PlotAssignment() PlotAssignmentRepository
	DeviceToken() DeviceTokenRepository
	NotificationLog() NotificationLogRepository

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
