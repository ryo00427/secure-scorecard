package repository

import (
	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// Repository provides access to the database
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new Repository instance
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB returns the underlying database connection
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// --- User Repository Methods ---

// CreateUser creates a new user
func (r *Repository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByFirebaseUID retrieves a user by Firebase UID
func (r *Repository) GetUserByFirebaseUID(uid string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("firebase_uid = ?", uid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user
func (r *Repository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

// --- Garden Repository Methods ---

// CreateGarden creates a new garden
func (r *Repository) CreateGarden(garden *model.Garden) error {
	return r.db.Create(garden).Error
}

// GetGardenByID retrieves a garden by ID
func (r *Repository) GetGardenByID(id uint) (*model.Garden, error) {
	var garden model.Garden
	if err := r.db.Preload("User").First(&garden, id).Error; err != nil {
		return nil, err
	}
	return &garden, nil
}

// GetGardensByUserID retrieves all gardens for a user
func (r *Repository) GetGardensByUserID(userID uint) ([]model.Garden, error) {
	var gardens []model.Garden
	if err := r.db.Where("user_id = ?", userID).Find(&gardens).Error; err != nil {
		return nil, err
	}
	return gardens, nil
}

// UpdateGarden updates a garden
func (r *Repository) UpdateGarden(garden *model.Garden) error {
	return r.db.Save(garden).Error
}

// DeleteGarden soft deletes a garden
func (r *Repository) DeleteGarden(id uint) error {
	return r.db.Delete(&model.Garden{}, id).Error
}

// --- Plant Repository Methods ---

// CreatePlant creates a new plant
func (r *Repository) CreatePlant(plant *model.Plant) error {
	return r.db.Create(plant).Error
}

// GetPlantByID retrieves a plant by ID
func (r *Repository) GetPlantByID(id uint) (*model.Plant, error) {
	var plant model.Plant
	if err := r.db.Preload("Garden").First(&plant, id).Error; err != nil {
		return nil, err
	}
	return &plant, nil
}

// GetPlantsByGardenID retrieves all plants for a garden
func (r *Repository) GetPlantsByGardenID(gardenID uint) ([]model.Plant, error) {
	var plants []model.Plant
	if err := r.db.Where("garden_id = ?", gardenID).Find(&plants).Error; err != nil {
		return nil, err
	}
	return plants, nil
}

// UpdatePlant updates a plant
func (r *Repository) UpdatePlant(plant *model.Plant) error {
	return r.db.Save(plant).Error
}

// DeletePlant soft deletes a plant
func (r *Repository) DeletePlant(id uint) error {
	return r.db.Delete(&model.Plant{}, id).Error
}

// --- CareLog Repository Methods ---

// CreateCareLog creates a new care log
func (r *Repository) CreateCareLog(careLog *model.CareLog) error {
	return r.db.Create(careLog).Error
}

// GetCareLogsByPlantID retrieves all care logs for a plant
func (r *Repository) GetCareLogsByPlantID(plantID uint) ([]model.CareLog, error) {
	var careLogs []model.CareLog
	if err := r.db.Where("plant_id = ?", plantID).Order("cared_at DESC").Find(&careLogs).Error; err != nil {
		return nil, err
	}
	return careLogs, nil
}
