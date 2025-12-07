package repository

import (
	"context"
	"time"

	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// =============================================================================
// CropRepository Implementation - 作物リポジトリ
// =============================================================================

// cropRepository implements CropRepository
type cropRepository struct {
	db *gorm.DB
}

// Create creates a new crop
func (r *cropRepository) Create(ctx context.Context, crop *model.Crop) error {
	return GetDB(ctx, r.db).Create(crop).Error
}

// GetByID retrieves a crop by ID
func (r *cropRepository) GetByID(ctx context.Context, id uint) (*model.Crop, error) {
	var crop model.Crop
	if err := GetDB(ctx, r.db).First(&crop, id).Error; err != nil {
		return nil, err
	}
	return &crop, nil
}

// GetByUserID retrieves all crops for a user
func (r *cropRepository) GetByUserID(ctx context.Context, userID uint) ([]model.Crop, error) {
	var crops []model.Crop
	if err := GetDB(ctx, r.db).Where("user_id = ?", userID).Order("planted_date DESC").Find(&crops).Error; err != nil {
		return nil, err
	}
	return crops, nil
}

// GetByUserIDAndStatus retrieves crops for a user with a specific status
func (r *cropRepository) GetByUserIDAndStatus(ctx context.Context, userID uint, status string) ([]model.Crop, error) {
	var crops []model.Crop
	if err := GetDB(ctx, r.db).Where("user_id = ? AND status = ?", userID, status).Order("planted_date DESC").Find(&crops).Error; err != nil {
		return nil, err
	}
	return crops, nil
}

// GetUpcomingHarvests は指定日数以内に収穫予定の作物を取得します（通知処理用）
// ユーザー情報を含めて取得し、収穫リマインダー通知に使用します
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - daysAhead: 今日から何日後までを対象とするか
//
// 戻り値:
//   - []model.Crop: 収穫予定の作物一覧（ユーザー情報を含む）
//   - error: 取得に失敗した場合のエラー
func (r *cropRepository) GetUpcomingHarvests(ctx context.Context, daysAhead int) ([]model.Crop, error) {
	var crops []model.Crop
	today := time.Now().Truncate(24 * time.Hour)
	targetDate := today.AddDate(0, 0, daysAhead)

	// status が "growing" で、expected_harvest_date が今日から daysAhead 日後の作物を取得
	if err := GetDB(ctx, r.db).
		Preload("User").
		Where("status = ? AND expected_harvest_date >= ? AND expected_harvest_date <= ?",
			"growing", today, targetDate).
		Order("user_id ASC, expected_harvest_date ASC").
		Find(&crops).Error; err != nil {
		return nil, err
	}
	return crops, nil
}

// Update updates a crop
func (r *cropRepository) Update(ctx context.Context, crop *model.Crop) error {
	return GetDB(ctx, r.db).Save(crop).Error
}

// Delete soft deletes a crop
func (r *cropRepository) Delete(ctx context.Context, id uint) error {
	return GetDB(ctx, r.db).Delete(&model.Crop{}, id).Error
}

// =============================================================================
// GrowthRecordRepository Implementation - 成長記録リポジトリ
// =============================================================================

// growthRecordRepository implements GrowthRecordRepository
type growthRecordRepository struct {
	db *gorm.DB
}

// Create creates a new growth record
func (r *growthRecordRepository) Create(ctx context.Context, record *model.GrowthRecord) error {
	return GetDB(ctx, r.db).Create(record).Error
}

// GetByID retrieves a growth record by ID
func (r *growthRecordRepository) GetByID(ctx context.Context, id uint) (*model.GrowthRecord, error) {
	var record model.GrowthRecord
	if err := GetDB(ctx, r.db).First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetByCropID retrieves all growth records for a crop
func (r *growthRecordRepository) GetByCropID(ctx context.Context, cropID uint) ([]model.GrowthRecord, error) {
	var records []model.GrowthRecord
	if err := GetDB(ctx, r.db).Where("crop_id = ?", cropID).Order("record_date DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// Delete soft deletes a growth record
func (r *growthRecordRepository) Delete(ctx context.Context, id uint) error {
	return GetDB(ctx, r.db).Delete(&model.GrowthRecord{}, id).Error
}

// DeleteByCropID deletes all growth records for a crop (batch delete to avoid N+1)
func (r *growthRecordRepository) DeleteByCropID(ctx context.Context, cropID uint) error {
	return GetDB(ctx, r.db).Where("crop_id = ?", cropID).Delete(&model.GrowthRecord{}).Error
}

// =============================================================================
// HarvestRepository Implementation - 収穫記録リポジトリ
// =============================================================================

// harvestRepository implements HarvestRepository
type harvestRepository struct {
	db *gorm.DB
}

// Create creates a new harvest record
func (r *harvestRepository) Create(ctx context.Context, harvest *model.Harvest) error {
	return GetDB(ctx, r.db).Create(harvest).Error
}

// GetByID retrieves a harvest record by ID
func (r *harvestRepository) GetByID(ctx context.Context, id uint) (*model.Harvest, error) {
	var harvest model.Harvest
	if err := GetDB(ctx, r.db).First(&harvest, id).Error; err != nil {
		return nil, err
	}
	return &harvest, nil
}

// GetByCropID retrieves all harvest records for a crop
func (r *harvestRepository) GetByCropID(ctx context.Context, cropID uint) ([]model.Harvest, error) {
	var harvests []model.Harvest
	if err := GetDB(ctx, r.db).Where("crop_id = ?", cropID).Order("harvest_date DESC").Find(&harvests).Error; err != nil {
		return nil, err
	}
	return harvests, nil
}

// Delete soft deletes a harvest record
func (r *harvestRepository) Delete(ctx context.Context, id uint) error {
	return GetDB(ctx, r.db).Delete(&model.Harvest{}, id).Error
}

// DeleteByCropID deletes all harvest records for a crop (batch delete to avoid N+1)
func (r *harvestRepository) DeleteByCropID(ctx context.Context, cropID uint) error {
	return GetDB(ctx, r.db).Where("crop_id = ?", cropID).Delete(&model.Harvest{}).Error
}

// GetByUserIDWithDateRange はユーザーの収穫記録を日付範囲でフィルタして取得します。
// Analytics用のクエリで、cropsテーブルとJOINしてユーザーの収穫データを取得します。
// startDate/endDateがnilの場合は、その方向の制限はありません。
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - userID: ユーザーID
//   - startDate: 開始日（nilの場合は制限なし）
//   - endDate: 終了日（nilの場合は制限なし）
//
// 戻り値:
//   - []model.Harvest: 収穫記録の一覧（収穫日の降順）
//   - error: 取得に失敗した場合のエラー
func (r *harvestRepository) GetByUserIDWithDateRange(ctx context.Context, userID uint, startDate, endDate *time.Time) ([]model.Harvest, error) {
	db := GetDB(ctx, r.db)

	// cropsテーブルとJOINしてユーザーの収穫データを取得
	query := db.Joins("JOIN crops ON crops.id = harvests.crop_id AND crops.deleted_at IS NULL").
		Where("crops.user_id = ?", userID)

	// 日付範囲フィルタ
	if startDate != nil {
		query = query.Where("harvests.harvest_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("harvests.harvest_date <= ?", *endDate)
	}

	var harvests []model.Harvest
	if err := query.Order("harvests.harvest_date DESC").Find(&harvests).Error; err != nil {
		return nil, err
	}
	return harvests, nil
}
