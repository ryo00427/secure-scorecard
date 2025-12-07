// Package database - データベース接続とマイグレーション管理
//
// 機能:
//   - PostgreSQL接続管理（接続プール設定含む）
//   - GORMマイグレーション
//   - カスタムインデックスと制約の作成
//   - Materialized Viewの作成と管理
//   - ヘルスチェック
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

// =============================================================================
// マイグレーション
// =============================================================================

// AutoMigrate runs database migrations for all models
// すべてのモデルに対してGORMの自動マイグレーションを実行します。
func (db *DB) AutoMigrate() error {
	log.Println("Running database migrations...")

	// すべてのモデルをマイグレーション
	if err := db.DB.AutoMigrate(
		// 認証・ユーザー関連
		&model.User{},
		&model.TokenBlacklist{},

		// 菜園・植物関連（レガシー）
		&model.Garden{},
		&model.Plant{},
		&model.CareLog{},

		// 作物管理
		&model.Crop{},
		&model.GrowthRecord{},
		&model.Harvest{},

		// 区画管理
		&model.Plot{},
		&model.PlotAssignment{},

		// タスク管理
		&model.Task{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// =============================================================================
// インデックス作成
// =============================================================================

// CreateIndexes creates additional indexes for performance optimization
// パフォーマンス最適化のためのカスタムインデックスを作成します。
func (db *DB) CreateIndexes() error {
	log.Println("Creating database indexes...")

	indexes := []string{
		// =================================================================
		// users テーブル
		// =================================================================
		// メール検索用（ログイン時）
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		// Firebase UID検索用
		`CREATE INDEX IF NOT EXISTS idx_users_firebase_uid ON users(firebase_uid) WHERE firebase_uid IS NOT NULL`,
		// アクティブユーザー検索用
		`CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active) WHERE is_active = true`,

		// =================================================================
		// token_blacklist テーブル
		// =================================================================
		// トークンハッシュ検索用（認証時）
		`CREATE INDEX IF NOT EXISTS idx_token_blacklist_token_hash ON token_blacklist(token_hash)`,
		// 期限切れトークン削除用
		`CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires_at ON token_blacklist(expires_at)`,

		// =================================================================
		// crops テーブル
		// =================================================================
		// ユーザー別作物一覧用
		`CREATE INDEX IF NOT EXISTS idx_crops_user_id ON crops(user_id)`,
		// ステータス別フィルタ用
		`CREATE INDEX IF NOT EXISTS idx_crops_status ON crops(status)`,
		// ユーザー×ステータス複合インデックス
		`CREATE INDEX IF NOT EXISTS idx_crops_user_status ON crops(user_id, status)`,
		// 収穫予定日検索用（リマインダー）
		`CREATE INDEX IF NOT EXISTS idx_crops_expected_harvest_date ON crops(expected_harvest_date)`,
		// 植え付け日でのソート用
		`CREATE INDEX IF NOT EXISTS idx_crops_planted_date ON crops(planted_date)`,

		// =================================================================
		// growth_records テーブル
		// =================================================================
		// 作物別成長記録取得用
		`CREATE INDEX IF NOT EXISTS idx_growth_records_crop_id ON growth_records(crop_id)`,
		// 記録日でのソート用
		`CREATE INDEX IF NOT EXISTS idx_growth_records_record_date ON growth_records(record_date)`,
		// 作物×記録日複合インデックス
		`CREATE INDEX IF NOT EXISTS idx_growth_records_crop_date ON growth_records(crop_id, record_date DESC)`,

		// =================================================================
		// harvests テーブル
		// =================================================================
		// 作物別収穫記録取得用
		`CREATE INDEX IF NOT EXISTS idx_harvests_crop_id ON harvests(crop_id)`,
		// 収穫日でのソート用
		`CREATE INDEX IF NOT EXISTS idx_harvests_harvest_date ON harvests(harvest_date)`,
		// 分析用: 期間指定での集計
		`CREATE INDEX IF NOT EXISTS idx_harvests_crop_date ON harvests(crop_id, harvest_date)`,

		// =================================================================
		// plots テーブル
		// =================================================================
		// ユーザー別区画一覧用
		`CREATE INDEX IF NOT EXISTS idx_plots_user_id ON plots(user_id)`,
		// ステータス別フィルタ用
		`CREATE INDEX IF NOT EXISTS idx_plots_status ON plots(status)`,
		// ユーザー×ステータス複合インデックス
		`CREATE INDEX IF NOT EXISTS idx_plots_user_status ON plots(user_id, status)`,

		// =================================================================
		// plot_assignments テーブル
		// =================================================================
		// 区画別配置履歴取得用
		`CREATE INDEX IF NOT EXISTS idx_plot_assignments_plot_id ON plot_assignments(plot_id)`,
		// 作物別配置履歴取得用
		`CREATE INDEX IF NOT EXISTS idx_plot_assignments_crop_id ON plot_assignments(crop_id)`,
		// アクティブな配置検索用
		`CREATE INDEX IF NOT EXISTS idx_plot_assignments_active ON plot_assignments(plot_id) WHERE unassigned_date IS NULL`,

		// =================================================================
		// tasks テーブル
		// =================================================================
		// ユーザー別タスク一覧用
		`CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id)`,
		// ステータス別フィルタ用
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		// 期限日でのソート用
		`CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date)`,
		// 今日のタスク検索用
		`CREATE INDEX IF NOT EXISTS idx_tasks_user_due_date ON tasks(user_id, due_date)`,
		// 期限切れタスク検索用
		`CREATE INDEX IF NOT EXISTS idx_tasks_overdue ON tasks(user_id, due_date, status) WHERE status = 'pending'`,
		// 繰り返しタスク検索用
		`CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_task_id) WHERE parent_task_id IS NOT NULL`,
	}

	for _, idx := range indexes {
		if err := db.DB.Exec(idx).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
			// インデックス作成失敗は警告のみ（既存インデックスの場合もある）
		}
	}

	log.Println("Database indexes created successfully")
	return nil
}

// =============================================================================
// 制約作成
// =============================================================================

// CreateConstraints creates CHECK constraints and additional database constraints
// CHECK制約やその他のデータベース制約を作成します。
func (db *DB) CreateConstraints() error {
	log.Println("Creating database constraints...")

	constraints := []string{
		// =================================================================
		// crops テーブル - 日付バリデーション
		// =================================================================
		// 植え付け日 <= 収穫予定日
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_crops_valid_dates'
			) THEN
				ALTER TABLE crops ADD CONSTRAINT chk_crops_valid_dates
					CHECK (planted_date <= expected_harvest_date);
			END IF;
		END $$`,

		// ステータス値の制限
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_crops_status'
			) THEN
				ALTER TABLE crops ADD CONSTRAINT chk_crops_status
					CHECK (status IN ('planted', 'growing', 'ready_to_harvest', 'harvested', 'failed'));
			END IF;
		END $$`,

		// =================================================================
		// growth_records テーブル - 成長段階バリデーション
		// =================================================================
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_growth_records_stage'
			) THEN
				ALTER TABLE growth_records ADD CONSTRAINT chk_growth_records_stage
					CHECK (growth_stage IN ('seedling', 'vegetative', 'flowering', 'fruiting'));
			END IF;
		END $$`,

		// =================================================================
		// harvests テーブル - 数量バリデーション
		// =================================================================
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_harvests_quantity'
			) THEN
				ALTER TABLE harvests ADD CONSTRAINT chk_harvests_quantity
					CHECK (quantity > 0);
			END IF;
		END $$`,

		// 品質値の制限
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_harvests_quality'
			) THEN
				ALTER TABLE harvests ADD CONSTRAINT chk_harvests_quality
					CHECK (quality IS NULL OR quality IN ('excellent', 'good', 'fair', 'poor'));
			END IF;
		END $$`,

		// =================================================================
		// plots テーブル - サイズバリデーション
		// =================================================================
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_plots_dimensions'
			) THEN
				ALTER TABLE plots ADD CONSTRAINT chk_plots_dimensions
					CHECK (width > 0 AND height > 0);
			END IF;
		END $$`,

		// 土壌タイプの制限
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_plots_soil_type'
			) THEN
				ALTER TABLE plots ADD CONSTRAINT chk_plots_soil_type
					CHECK (soil_type IS NULL OR soil_type IN ('clay', 'sandy', 'loamy', 'peaty'));
			END IF;
		END $$`,

		// 日当たりの制限
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_plots_sunlight'
			) THEN
				ALTER TABLE plots ADD CONSTRAINT chk_plots_sunlight
					CHECK (sunlight IS NULL OR sunlight IN ('full_sun', 'partial_shade', 'shade'));
			END IF;
		END $$`,

		// ステータスの制限
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_plots_status'
			) THEN
				ALTER TABLE plots ADD CONSTRAINT chk_plots_status
					CHECK (status IN ('available', 'occupied'));
			END IF;
		END $$`,

		// =================================================================
		// tasks テーブル - ステータス/優先度バリデーション
		// =================================================================
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_tasks_status'
			) THEN
				ALTER TABLE tasks ADD CONSTRAINT chk_tasks_status
					CHECK (status IN ('pending', 'completed', 'cancelled'));
			END IF;
		END $$`,

		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_tasks_priority'
			) THEN
				ALTER TABLE tasks ADD CONSTRAINT chk_tasks_priority
					CHECK (priority IN ('low', 'medium', 'high'));
			END IF;
		END $$`,

		// 繰り返し設定の制限
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_tasks_recurrence'
			) THEN
				ALTER TABLE tasks ADD CONSTRAINT chk_tasks_recurrence
					CHECK (recurrence IS NULL OR recurrence = '' OR recurrence IN ('daily', 'weekly', 'monthly'));
			END IF;
		END $$`,

		// 繰り返し間隔は正の値
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_tasks_recurrence_interval'
			) THEN
				ALTER TABLE tasks ADD CONSTRAINT chk_tasks_recurrence_interval
					CHECK (recurrence_interval > 0);
			END IF;
		END $$`,
	}

	for _, constraint := range constraints {
		if err := db.DB.Exec(constraint).Error; err != nil {
			log.Printf("Warning: Failed to create constraint: %v", err)
		}
	}

	log.Println("Database constraints created successfully")
	return nil
}

// =============================================================================
// Materialized View
// =============================================================================

// CreateMaterializedViews creates materialized views for analytics
// 分析用のマテリアライズドビューを作成します。
func (db *DB) CreateMaterializedViews() error {
	log.Println("Creating materialized views...")

	// 収穫分析用マテリアライズドビュー
	mvHarvestAnalytics := `
	CREATE MATERIALIZED VIEW IF NOT EXISTS mv_harvest_analytics AS
	SELECT
		c.user_id,
		c.id as crop_id,
		c.name as crop_name,
		c.variety,
		c.planted_date,
		COALESCE(h.total_quantity, 0) as total_quantity,
		h.quantity_unit,
		h.harvest_count,
		h.first_harvest_date,
		h.last_harvest_date,
		CASE
			WHEN h.first_harvest_date IS NOT NULL
			THEN h.first_harvest_date - c.planted_date
			ELSE NULL
		END as days_to_first_harvest,
		h.avg_quality_score,
		c.status
	FROM crops c
	LEFT JOIN (
		SELECT
			crop_id,
			SUM(quantity) as total_quantity,
			MAX(quantity_unit) as quantity_unit,
			COUNT(*) as harvest_count,
			MIN(harvest_date) as first_harvest_date,
			MAX(harvest_date) as last_harvest_date,
			AVG(
				CASE quality
					WHEN 'excellent' THEN 4
					WHEN 'good' THEN 3
					WHEN 'fair' THEN 2
					WHEN 'poor' THEN 1
					ELSE NULL
				END
			) as avg_quality_score
		FROM harvests
		WHERE deleted_at IS NULL
		GROUP BY crop_id
	) h ON c.id = h.crop_id
	WHERE c.deleted_at IS NULL
	`

	if err := db.DB.Exec(mvHarvestAnalytics).Error; err != nil {
		log.Printf("Warning: Failed to create mv_harvest_analytics: %v", err)
	}

	// マテリアライズドビュー用インデックス
	mvIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_mv_harvest_analytics_user_id ON mv_harvest_analytics(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mv_harvest_analytics_crop_id ON mv_harvest_analytics(crop_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mv_harvest_analytics_planted_date ON mv_harvest_analytics(planted_date)`,
	}

	for _, idx := range mvIndexes {
		if err := db.DB.Exec(idx).Error; err != nil {
			log.Printf("Warning: Failed to create materialized view index: %v", err)
		}
	}

	// 月別収穫集計ビュー
	mvMonthlyHarvest := `
	CREATE MATERIALIZED VIEW IF NOT EXISTS mv_monthly_harvest AS
	SELECT
		c.user_id,
		DATE_TRUNC('month', h.harvest_date) as harvest_month,
		c.name as crop_name,
		SUM(h.quantity) as total_quantity,
		MAX(h.quantity_unit) as quantity_unit,
		COUNT(*) as harvest_count
	FROM harvests h
	JOIN crops c ON h.crop_id = c.id
	WHERE h.deleted_at IS NULL AND c.deleted_at IS NULL
	GROUP BY c.user_id, DATE_TRUNC('month', h.harvest_date), c.name
	`

	if err := db.DB.Exec(mvMonthlyHarvest).Error; err != nil {
		log.Printf("Warning: Failed to create mv_monthly_harvest: %v", err)
	}

	// 月別収穫集計ビュー用インデックス
	mvMonthlyIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_mv_monthly_harvest_user_id ON mv_monthly_harvest(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mv_monthly_harvest_month ON mv_monthly_harvest(harvest_month)`,
	}

	for _, idx := range mvMonthlyIndexes {
		if err := db.DB.Exec(idx).Error; err != nil {
			log.Printf("Warning: Failed to create monthly harvest index: %v", err)
		}
	}

	log.Println("Materialized views created successfully")
	return nil
}

// RefreshMaterializedViews refreshes all materialized views
// すべてのマテリアライズドビューをリフレッシュします。
// 通常は日次のcronジョブから呼び出されます。
func (db *DB) RefreshMaterializedViews() error {
	log.Println("Refreshing materialized views...")

	views := []string{
		"mv_harvest_analytics",
		"mv_monthly_harvest",
	}

	for _, view := range views {
		// CONCURRENTLY オプションを使用してロックを最小化
		sql := fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", view)
		if err := db.DB.Exec(sql).Error; err != nil {
			// CONCURRENTLY が失敗した場合は通常のリフレッシュを試行
			sqlNormal := fmt.Sprintf("REFRESH MATERIALIZED VIEW %s", view)
			if err := db.DB.Exec(sqlNormal).Error; err != nil {
				log.Printf("Warning: Failed to refresh %s: %v", view, err)
			}
		}
	}

	log.Println("Materialized views refreshed successfully")
	return nil
}

// =============================================================================
// 期限切れデータのクリーンアップ
// =============================================================================

// CleanupExpiredTokens removes expired tokens from the blacklist
// 期限切れのトークンをブラックリストから削除します。
// 通常は日次のcronジョブから呼び出されます。
func (db *DB) CleanupExpiredTokens() (int64, error) {
	log.Println("Cleaning up expired tokens...")

	result := db.DB.Exec("DELETE FROM token_blacklist WHERE expires_at < NOW()")
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", result.Error)
	}

	log.Printf("Cleaned up %d expired tokens", result.RowsAffected)
	return result.RowsAffected, nil
}

// =============================================================================
// 完全セットアップ
// =============================================================================

// Setup runs all database setup tasks
// データベースの完全セットアップを実行します（マイグレーション、インデックス、制約、ビュー）。
func (db *DB) Setup() error {
	if err := db.AutoMigrate(); err != nil {
		return err
	}

	if err := db.CreateIndexes(); err != nil {
		return err
	}

	if err := db.CreateConstraints(); err != nil {
		return err
	}

	if err := db.CreateMaterializedViews(); err != nil {
		return err
	}

	return nil
}
