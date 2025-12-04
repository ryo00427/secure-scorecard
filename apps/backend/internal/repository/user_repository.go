package repository

import (
	"context"

	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/gorm"
)

// userRepository implements UserRepository
type userRepository struct {
	db *gorm.DB
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return getDB(ctx, r.db).Create(user).Error
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := getDB(ctx, r.db).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByFirebaseUID retrieves a user by Firebase UID
func (r *userRepository) GetByFirebaseUID(ctx context.Context, uid string) (*model.User, error) {
	var user model.User
	if err := getDB(ctx, r.db).Where("firebase_uid = ?", uid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := getDB(ctx, r.db).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return getDB(ctx, r.db).Save(user).Error
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return getDB(ctx, r.db).Delete(&model.User{}, id).Error
}
