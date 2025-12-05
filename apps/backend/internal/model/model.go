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
