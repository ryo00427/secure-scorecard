package repository

import (
	"context"

	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// =============================================================================
// PlotRepository - 区画リポジトリ実装
// =============================================================================

// plotRepository implements PlotRepository interface
// 菜園の区画データへのアクセスを提供します
type plotRepository struct {
	db *gorm.DB
}

// Create は新しい区画を作成します
func (r *plotRepository) Create(ctx context.Context, plot *model.Plot) error {
	db := GetDB(ctx, r.db)
	return db.Create(plot).Error
}

// GetByID は指定されたIDの区画を取得します
func (r *plotRepository) GetByID(ctx context.Context, id uint) (*model.Plot, error) {
	db := GetDB(ctx, r.db)
	var plot model.Plot
	if err := db.First(&plot, id).Error; err != nil {
		return nil, err
	}
	return &plot, nil
}

// GetByUserID は指定されたユーザーの全区画を取得します
func (r *plotRepository) GetByUserID(ctx context.Context, userID uint) ([]model.Plot, error) {
	db := GetDB(ctx, r.db)
	var plots []model.Plot
	if err := db.Where("user_id = ?", userID).Find(&plots).Error; err != nil {
		return nil, err
	}
	return plots, nil
}

// GetByUserIDAndStatus は指定されたユーザーの特定ステータスの区画を取得します
// ステータス: available（空き）, occupied（使用中）
func (r *plotRepository) GetByUserIDAndStatus(ctx context.Context, userID uint, status string) ([]model.Plot, error) {
	db := GetDB(ctx, r.db)
	var plots []model.Plot
	if err := db.Where("user_id = ? AND status = ?", userID, status).Find(&plots).Error; err != nil {
		return nil, err
	}
	return plots, nil
}

// Update は区画情報を更新します
func (r *plotRepository) Update(ctx context.Context, plot *model.Plot) error {
	db := GetDB(ctx, r.db)
	return db.Save(plot).Error
}

// Delete は区画を削除します（ソフトデリート）
func (r *plotRepository) Delete(ctx context.Context, id uint) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&model.Plot{}, id).Error
}

// =============================================================================
// PlotAssignmentRepository - 区画配置リポジトリ実装
// =============================================================================

// plotAssignmentRepository implements PlotAssignmentRepository interface
// 区画への作物配置データへのアクセスを提供します
type plotAssignmentRepository struct {
	db *gorm.DB
}

// Create は新しい区画配置を作成します
func (r *plotAssignmentRepository) Create(ctx context.Context, assignment *model.PlotAssignment) error {
	db := GetDB(ctx, r.db)
	return db.Create(assignment).Error
}

// GetByID は指定されたIDの区画配置を取得します
func (r *plotAssignmentRepository) GetByID(ctx context.Context, id uint) (*model.PlotAssignment, error) {
	db := GetDB(ctx, r.db)
	var assignment model.PlotAssignment
	if err := db.First(&assignment, id).Error; err != nil {
		return nil, err
	}
	return &assignment, nil
}

// GetByPlotID は指定された区画の全配置履歴を取得します
func (r *plotAssignmentRepository) GetByPlotID(ctx context.Context, plotID uint) ([]model.PlotAssignment, error) {
	db := GetDB(ctx, r.db)
	var assignments []model.PlotAssignment
	if err := db.Where("plot_id = ?", plotID).Order("assigned_date DESC").Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

// GetActiveByPlotID は指定された区画の現在アクティブな配置を取得します
// アクティブ = UnassignedDate が NULL
func (r *plotAssignmentRepository) GetActiveByPlotID(ctx context.Context, plotID uint) (*model.PlotAssignment, error) {
	db := GetDB(ctx, r.db)
	var assignment model.PlotAssignment
	if err := db.Where("plot_id = ? AND unassigned_date IS NULL", plotID).First(&assignment).Error; err != nil {
		return nil, err
	}
	return &assignment, nil
}

// GetByCropID は指定された作物の全配置履歴を取得します
func (r *plotAssignmentRepository) GetByCropID(ctx context.Context, cropID uint) ([]model.PlotAssignment, error) {
	db := GetDB(ctx, r.db)
	var assignments []model.PlotAssignment
	if err := db.Where("crop_id = ?", cropID).Order("assigned_date DESC").Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

// Update は区画配置情報を更新します
func (r *plotAssignmentRepository) Update(ctx context.Context, assignment *model.PlotAssignment) error {
	db := GetDB(ctx, r.db)
	return db.Save(assignment).Error
}

// Delete は区画配置を削除します（ソフトデリート）
func (r *plotAssignmentRepository) Delete(ctx context.Context, id uint) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&model.PlotAssignment{}, id).Error
}

// DeleteByPlotID は指定された区画の全配置を削除します（バッチ削除）
// N+1問題を回避するため、一括削除を使用
func (r *plotAssignmentRepository) DeleteByPlotID(ctx context.Context, plotID uint) error {
	db := GetDB(ctx, r.db)
	return db.Where("plot_id = ?", plotID).Delete(&model.PlotAssignment{}).Error
}
