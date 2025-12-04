package database

import (
	"fmt"
	"log"
	"time"

	"github.com/secure-scorecard/backend/internal/config"
	"github.com/secure-scorecard/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB holds the database connection
type DB struct {
	*gorm.DB
}

// Config holds database connection configuration
type Config struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns the default database configuration
func DefaultConfig() *Config {
	return &Config{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// Connect establishes a database connection
func Connect(cfg *config.Config, dbCfg *Config) (*DB, error) {
	if dbCfg == nil {
		dbCfg = DefaultConfig()
	}

	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	if cfg.Server.Env == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true, // Performance: disable default transaction for single operations
		PrepareStmt:            true, // Performance: cache prepared statements
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(dbCfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(dbCfg.ConnMaxIdleTime)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connected successfully (pool: idle=%d, open=%d)",
		dbCfg.MaxIdleConns, dbCfg.MaxOpenConns)

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// HealthCheck checks if the database connection is healthy
func (db *DB) HealthCheck() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Stats returns database connection pool statistics
func (db *DB) Stats() map[string]interface{} {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
	}
}

// AutoMigrate runs database migrations
func (db *DB) AutoMigrate() error {
	log.Println("Running database migrations...")

	if err := db.DB.AutoMigrate(
		&model.User{},
		&model.Garden{},
		&model.Plant{},
		&model.CareLog{},
		&model.TokenBlacklist{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// CreateIndexes creates additional indexes for performance
func (db *DB) CreateIndexes() error {
	log.Println("Creating database indexes...")

	// Add any custom indexes here
	// Example:
	// db.DB.Exec("CREATE INDEX IF NOT EXISTS idx_plants_status ON plants(status)")

	log.Println("Database indexes created successfully")
	return nil
}
