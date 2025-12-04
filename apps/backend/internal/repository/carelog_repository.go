package repository

import (
	"context"

	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// careLogRepository implements CareLogRepository
type careLogRepository struct {
	db *gorm.DB
}

// Create creates a new care log
func (r *careLogRepository) Create(ctx context.Context, careLog *model.CareLog) error {
	return GetDB(ctx, r.db).Create(careLog).Error
}

// GetByID retrieves a care log by ID
func (r *careLogRepository) GetByID(ctx context.Context, id uint) (*model.CareLog, error) {
	var careLog model.CareLog
	if err := GetDB(ctx, r.db).Preload("Plant").First(&careLog, id).Error; err != nil {
		return nil, err
	}
	return &careLog, nil
}

// GetByPlantID retrieves all care logs for a plant
func (r *careLogRepository) GetByPlantID(ctx context.Context, plantID uint) ([]model.CareLog, error) {
	var careLogs []model.CareLog
	if err := GetDB(ctx, r.db).Where("plant_id = ?", plantID).Order("cared_at DESC").Find(&careLogs).Error; err != nil {
		return nil, err
	}
	return careLogs, nil
}

// Delete soft deletes a care log
func (r *careLogRepository) Delete(ctx context.Context, id uint) error {
	return GetDB(ctx, r.db).Delete(&model.CareLog{}, id).Error
}
