package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// txKey is the context key for storing transaction
type txKey struct{}

// TxFromContext retrieves the transaction from context
func TxFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil
}

// ContextWithTx returns a new context with the transaction
func ContextWithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// repositoryManager implements Repositories interface with transaction support
type repositoryManager struct {
	db              *gorm.DB
	user            *userRepository
	garden          *gardenRepository
	plant           *plantRepository
	careLog         *careLogRepository
	tokenBlacklist  *tokenBlacklistRepository
	task            *taskRepository
	crop            *cropRepository
	growthRecord    *growthRecordRepository
	harvest         *harvestRepository
	plot            *plotRepository
	plotAssignment  *plotAssignmentRepository
	deviceToken     *deviceTokenRepository
	notificationLog *notificationLogRepository
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(db *gorm.DB) Repositories {
	return &repositoryManager{
		db:              db,
		user:            &userRepository{db: db},
		garden:          &gardenRepository{db: db},
		plant:           &plantRepository{db: db},
		careLog:         &careLogRepository{db: db},
		tokenBlacklist:  &tokenBlacklistRepository{db: db},
		task:            &taskRepository{db: db},
		crop:            &cropRepository{db: db},
		growthRecord:    &growthRecordRepository{db: db},
		harvest:         &harvestRepository{db: db},
		plot:            &plotRepository{db: db},
		plotAssignment:  &plotAssignmentRepository{db: db},
		deviceToken:     &deviceTokenRepository{db: db},
		notificationLog: &notificationLogRepository{db: db},
	}
}

// User returns the user repository
func (m *repositoryManager) User() UserRepository {
	return m.user
}

// Garden returns the garden repository
func (m *repositoryManager) Garden() GardenRepository {
	return m.garden
}

// Plant returns the plant repository
func (m *repositoryManager) Plant() PlantRepository {
	return m.plant
}

// CareLog returns the care log repository
func (m *repositoryManager) CareLog() CareLogRepository {
	return m.careLog
}

// TokenBlacklist returns the token blacklist repository
func (m *repositoryManager) TokenBlacklist() TokenBlacklistRepository {
	return m.tokenBlacklist
}

// Task returns the task repository
func (m *repositoryManager) Task() TaskRepository {
	return m.task
}

// Crop returns the crop repository
func (m *repositoryManager) Crop() CropRepository {
	return m.crop
}

// GrowthRecord returns the growth record repository
func (m *repositoryManager) GrowthRecord() GrowthRecordRepository {
	return m.growthRecord
}

// Harvest returns the harvest repository
func (m *repositoryManager) Harvest() HarvestRepository {
	return m.harvest
}

// Plot returns the plot repository
func (m *repositoryManager) Plot() PlotRepository {
	return m.plot
}

// PlotAssignment returns the plot assignment repository
func (m *repositoryManager) PlotAssignment() PlotAssignmentRepository {
	return m.plotAssignment
}

// DeviceToken returns the device token repository
func (m *repositoryManager) DeviceToken() DeviceTokenRepository {
	return m.deviceToken
}

// NotificationLog returns the notification log repository
func (m *repositoryManager) NotificationLog() NotificationLogRepository {
	return m.notificationLog
}

// WithTransaction executes a function within a database transaction
func (m *repositoryManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Check if already in a transaction
	if TxFromContext(ctx) != nil {
		return fn(ctx)
	}

	// Start new transaction
	tx := m.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Create context with transaction
	txCtx := ContextWithTx(ctx, tx)

	// Execute function
	if err := fn(txCtx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDB returns the appropriate database connection (transaction or main)
func GetDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return db.WithContext(ctx)
}
