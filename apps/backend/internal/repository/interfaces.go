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

// Repositories aggregates all repository interfaces
type Repositories interface {
	User() UserRepository
	Garden() GardenRepository
	Plant() PlantRepository
	CareLog() CareLogRepository
	TokenBlacklist() TokenBlacklistRepository

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
