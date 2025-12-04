package repository

import (
	"context"

	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// gardenRepository implements GardenRepository
type gardenRepository struct {
	db *gorm.DB
}

// Create creates a new garden
func (r *gardenRepository) Create(ctx context.Context, garden *model.Garden) error {
	return getDB(ctx, r.db).Create(garden).Error
}

// GetByID retrieves a garden by ID
func (r *gardenRepository) GetByID(ctx context.Context, id uint) (*model.Garden, error) {
	var garden model.Garden
	if err := getDB(ctx, r.db).Preload("User").First(&garden, id).Error; err != nil {
		return nil, err
	}
	return &garden, nil
}

// GetByUserID retrieves all gardens for a user
func (r *gardenRepository) GetByUserID(ctx context.Context, userID uint) ([]model.Garden, error) {
	var gardens []model.Garden
	if err := getDB(ctx, r.db).Where("user_id = ?", userID).Find(&gardens).Error; err != nil {
		return nil, err
	}
	return gardens, nil
}

// Update updates a garden
func (r *gardenRepository) Update(ctx context.Context, garden *model.Garden) error {
	return getDB(ctx, r.db).Save(garden).Error
}

// Delete soft deletes a garden
func (r *gardenRepository) Delete(ctx context.Context, id uint) error {
	return getDB(ctx, r.db).Delete(&model.Garden{}, id).Error
}
