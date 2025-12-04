package repository

import (
	"context"

	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// plantRepository implements PlantRepository
type plantRepository struct {
	db *gorm.DB
}

// Create creates a new plant
func (r *plantRepository) Create(ctx context.Context, plant *model.Plant) error {
	return getDB(ctx, r.db).Create(plant).Error
}

// GetByID retrieves a plant by ID
func (r *plantRepository) GetByID(ctx context.Context, id uint) (*model.Plant, error) {
	var plant model.Plant
	if err := getDB(ctx, r.db).Preload("Garden").First(&plant, id).Error; err != nil {
		return nil, err
	}
	return &plant, nil
}

// GetByGardenID retrieves all plants for a garden
func (r *plantRepository) GetByGardenID(ctx context.Context, gardenID uint) ([]model.Plant, error) {
	var plants []model.Plant
	if err := getDB(ctx, r.db).Where("garden_id = ?", gardenID).Find(&plants).Error; err != nil {
		return nil, err
	}
	return plants, nil
}

// Update updates a plant
func (r *plantRepository) Update(ctx context.Context, plant *model.Plant) error {
	return getDB(ctx, r.db).Save(plant).Error
}

// Delete soft deletes a plant
func (r *plantRepository) Delete(ctx context.Context, id uint) error {
	return getDB(ctx, r.db).Delete(&model.Plant{}, id).Error
}
