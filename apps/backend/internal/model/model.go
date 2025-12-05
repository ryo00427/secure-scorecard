package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// User represents a user in the system
type User struct {
	BaseModel
	FirebaseUID      string     `gorm:"uniqueIndex;size:128" json:"firebase_uid,omitempty"`
	Email            string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash     string     `gorm:"size:255" json:"-"`
	DisplayName      string     `gorm:"size:100" json:"display_name"`
	PhotoURL         string     `gorm:"size:500" json:"photo_url,omitempty"`
	IsActive         bool       `gorm:"default:true" json:"is_active"`
	FailedLoginCount int        `gorm:"default:0" json:"-"`
	LockedUntil      *time.Time `json:"-"`
}

// Garden represents a garden owned by a user
type Garden struct {
	BaseModel
	UserID      uint    `gorm:"index" json:"user_id"`
	Name        string  `gorm:"size:100;not null" json:"name"`
	Description string  `gorm:"size:500" json:"description,omitempty"`
	Location    string  `gorm:"size:200" json:"location,omitempty"`
	SizeM2      float64 `json:"size_m2,omitempty"`
	User        User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Plant represents a plant in a garden
type Plant struct {
	BaseModel
	GardenID     uint      `gorm:"index" json:"garden_id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Species      string    `gorm:"size:100" json:"species,omitempty"`
	PlantedAt    time.Time `json:"planted_at,omitempty"`
	HarvestedAt  time.Time `json:"harvested_at,omitempty"`
	Status       string    `gorm:"size:50;default:'growing'" json:"status"`
	Notes        string    `gorm:"size:1000" json:"notes,omitempty"`
	Garden       Garden    `gorm:"foreignKey:GardenID" json:"garden,omitempty"`
}

// CareLog represents a care activity for a plant
type CareLog struct {
	BaseModel
	PlantID uint      `gorm:"index" json:"plant_id"`
	Type    string    `gorm:"size:50;not null" json:"type"` // watering, fertilizing, pruning, etc.
	Notes   string    `gorm:"size:500" json:"notes,omitempty"`
	CaredAt time.Time `json:"cared_at"`
	Plant   Plant     `gorm:"foreignKey:PlantID" json:"plant,omitempty"`
}

// TokenBlacklist represents a blacklisted JWT token
type TokenBlacklist struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TokenHash string    `gorm:"uniqueIndex;size:64;not null" json:"token_hash"` // SHA-256 hash
	RevokedAt time.Time `gorm:"not null;index" json:"revoked_at"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
}

// Task represents a to-do task for gardening activities
// Task はタスク（やることリスト）を表すモデルです。
// 繰り返しタスクをサポートし、完了時に次回タスクを自動生成できます。
//
// 繰り返し設定:
//   - Recurrence: 繰り返し頻度（daily, weekly, monthly）
//   - RecurrenceInterval: 間隔（例: 2なら2日/2週/2ヶ月ごと）
//   - MaxOccurrences: 最大繰り返し回数（nilで無制限）
//   - RecurrenceEndDate: 繰り返し終了日（nilで無期限）
//   - OccurrenceCount: 現在の繰り返し回数
//   - ParentTaskID: 元タスクのID（繰り返しで生成されたタスクの場合）
type Task struct {
	BaseModel
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	PlantID     *uint      `gorm:"index" json:"plant_id,omitempty"` // Optional: link to specific plant
	Title       string     `gorm:"size:200;not null" json:"title"`
	Description string     `gorm:"size:1000" json:"description,omitempty"`
	DueDate     time.Time  `gorm:"index;not null" json:"due_date"`
	Priority    string     `gorm:"size:20;default:'medium'" json:"priority"` // low, medium, high
	Status      string     `gorm:"size:20;default:'pending'" json:"status"`  // pending, completed, cancelled
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// 繰り返し設定フィールド
	Recurrence         string     `gorm:"size:20" json:"recurrence,omitempty"`           // daily, weekly, monthly, or empty
	RecurrenceInterval int        `gorm:"default:1" json:"recurrence_interval,omitempty"` // every N days/weeks/months
	MaxOccurrences     *int       `json:"max_occurrences,omitempty"`                      // nil = unlimited
	RecurrenceEndDate  *time.Time `json:"recurrence_end_date,omitempty"`                  // nil = no end date
	OccurrenceCount    int        `gorm:"default:0" json:"occurrence_count"`              // current count
	ParentTaskID       *uint      `gorm:"index" json:"parent_task_id,omitempty"`          // original task ID

	// リレーション
	User       User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Plant      *Plant `gorm:"foreignKey:PlantID" json:"plant,omitempty"`
	ParentTask *Task  `gorm:"foreignKey:ParentTaskID" json:"parent_task,omitempty"`
}

// TableName overrides the table name for User
func (User) TableName() string {
	return "users"
}

// TableName overrides the table name for Garden
func (Garden) TableName() string {
	return "gardens"
}

// TableName overrides the table name for Plant
func (Plant) TableName() string {
	return "plants"
}

// TableName overrides the table name for CareLog
func (CareLog) TableName() string {
	return "care_logs"
}

// TableName overrides the table name for TokenBlacklist
func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

// TableName overrides the table name for Task
func (Task) TableName() string {
	return "tasks"
}

// =============================================================================
// Crop Domain Models - 作物管理モデル
// =============================================================================

// Crop は作物を表すモデルです。
// 植え付けから収穫までのライフサイクルを管理します。
//
// ステータス:
//   - planted: 植え付け済み
//   - growing: 成長中
//   - ready_to_harvest: 収穫可能
//   - harvested: 収穫済み
//   - failed: 失敗
//
// バリデーション:
//   - PlantedDate <= ExpectedHarvestDate
type Crop struct {
	BaseModel
	UserID              uint       `gorm:"index;not null" json:"user_id"`
	PlotID              *uint      `gorm:"index" json:"plot_id,omitempty"` // 区画への配置（任意）
	Name                string     `gorm:"size:100;not null" json:"name"`
	Variety             string     `gorm:"size:100" json:"variety,omitempty"` // 品種
	PlantedDate         time.Time  `gorm:"not null" json:"planted_date"`
	ExpectedHarvestDate time.Time  `gorm:"not null" json:"expected_harvest_date"`
	Status              string     `gorm:"size:20;default:'planted'" json:"status"` // planted, growing, ready_to_harvest, harvested, failed
	Notes               string     `gorm:"size:1000" json:"notes,omitempty"`

	// リレーション
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	GrowthRecords []GrowthRecord `gorm:"foreignKey:CropID" json:"growth_records,omitempty"`
	Harvests      []Harvest      `gorm:"foreignKey:CropID" json:"harvests,omitempty"`
}

// GrowthRecord は作物の成長記録を表すモデルです。
// 定期的な成長観察の記録を保存します。
//
// 成長段階:
//   - seedling: 苗
//   - vegetative: 成長期
//   - flowering: 開花期
//   - fruiting: 結実期
type GrowthRecord struct {
	BaseModel
	CropID      uint      `gorm:"index;not null" json:"crop_id"`
	RecordDate  time.Time `gorm:"not null" json:"record_date"`
	GrowthStage string    `gorm:"size:20;not null" json:"growth_stage"` // seedling, vegetative, flowering, fruiting
	Notes       string    `gorm:"size:1000" json:"notes,omitempty"`
	ImageURL    string    `gorm:"size:500" json:"image_url,omitempty"` // S3署名付きURL

	// リレーション
	Crop Crop `gorm:"foreignKey:CropID" json:"crop,omitempty"`
}

// Harvest は収穫記録を表すモデルです。
// 収穫量と品質を記録します。
//
// 品質評価:
//   - excellent: 優良
//   - good: 良好
//   - fair: 普通
//   - poor: 不良
type Harvest struct {
	BaseModel
	CropID       uint      `gorm:"index;not null" json:"crop_id"`
	HarvestDate  time.Time `gorm:"not null" json:"harvest_date"`
	Quantity     float64   `gorm:"not null" json:"quantity"`
	QuantityUnit string    `gorm:"size:20;not null" json:"quantity_unit"` // kg, g, pieces
	Quality      string    `gorm:"size:20" json:"quality,omitempty"`      // excellent, good, fair, poor
	Notes        string    `gorm:"size:1000" json:"notes,omitempty"`

	// リレーション
	Crop Crop `gorm:"foreignKey:CropID" json:"crop,omitempty"`
}

// TableName overrides the table name for Crop
func (Crop) TableName() string {
	return "crops"
}

// TableName overrides the table name for GrowthRecord
func (GrowthRecord) TableName() string {
	return "growth_records"
}

// TableName overrides the table name for Harvest
func (Harvest) TableName() string {
	return "harvests"
}

// =============================================================================
// Plot Domain Models - 区画管理モデル
// =============================================================================

// Plot は菜園の区画を表すモデルです。
// 菜園をグリッド状に分割し、作物の配置を管理します。
//
// ステータス:
//   - available: 空き
//   - occupied: 使用中
//
// 土壌タイプ:
//   - clay: 粘土質
//   - sandy: 砂質
//   - loamy: ローム（壌土）
//   - peaty: 泥炭質
//
// 日当たり:
//   - full_sun: 日向
//   - partial_shade: 半日陰
//   - shade: 日陰
//
// バリデーション:
//   - Width > 0, Height > 0
type Plot struct {
	BaseModel
	UserID    uint    `gorm:"index;not null" json:"user_id"`
	Name      string  `gorm:"size:100;not null" json:"name"`
	Width     float64 `gorm:"not null" json:"width"`            // メートル単位
	Height    float64 `gorm:"not null" json:"height"`           // メートル単位
	SoilType  string  `gorm:"size:20" json:"soil_type,omitempty"` // clay, sandy, loamy, peaty
	Sunlight  string  `gorm:"size:20" json:"sunlight,omitempty"`  // full_sun, partial_shade, shade
	Status    string  `gorm:"size:20;default:'available'" json:"status"` // available, occupied
	PositionX *int    `json:"position_x,omitempty"` // グリッド内のX座標（任意）
	PositionY *int    `json:"position_y,omitempty"` // グリッド内のY座標（任意）
	Notes     string  `gorm:"size:1000" json:"notes,omitempty"`

	// リレーション
	User            User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PlotAssignments []PlotAssignment  `gorm:"foreignKey:PlotID" json:"plot_assignments,omitempty"`
}

// PlotAssignment は区画への作物配置を表すモデルです。
// 区画と作物の関連付けを管理し、配置履歴を記録します。
type PlotAssignment struct {
	BaseModel
	PlotID       uint       `gorm:"index;not null" json:"plot_id"`
	CropID       uint       `gorm:"index;not null" json:"crop_id"`
	AssignedDate time.Time  `gorm:"not null" json:"assigned_date"`
	UnassignedDate *time.Time `json:"unassigned_date,omitempty"` // 配置解除日（履歴用）

	// リレーション
	Plot Plot `gorm:"foreignKey:PlotID" json:"plot,omitempty"`
	Crop Crop `gorm:"foreignKey:CropID" json:"crop,omitempty"`
}

// TableName overrides the table name for Plot
func (Plot) TableName() string {
	return "plots"
}

// TableName overrides the table name for PlotAssignment
func (PlotAssignment) TableName() string {
	return "plot_assignments"
}
