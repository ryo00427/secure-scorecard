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
	db      *gorm.DB
	user    *userRepository
	garden  *gardenRepository
	plant   *plantRepository
	careLog *careLogRepository
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(db *gorm.DB) Repositories {
	return &repositoryManager{
		db:      db,
		user:    &userRepository{db: db},
		garden:  &gardenRepository{db: db},
		plant:   &plantRepository{db: db},
		careLog: &careLogRepository{db: db},
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

// getDB returns the appropriate database connection (transaction or main)
func getDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return db.WithContext(ctx)
}
