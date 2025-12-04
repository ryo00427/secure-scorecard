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
	FirebaseUID string `gorm:"uniqueIndex;size:128" json:"firebase_uid"`
	Email       string `gorm:"uniqueIndex;size:255" json:"email"`
	DisplayName string `gorm:"size:100" json:"display_name"`
	PhotoURL    string `gorm:"size:500" json:"photo_url,omitempty"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
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
	PlantID   uint      `gorm:"index" json:"plant_id"`
	Type      string    `gorm:"size:50;not null" json:"type"` // watering, fertilizing, pruning, etc.
	Notes     string    `gorm:"size:500" json:"notes,omitempty"`
	CaredAt   time.Time `json:"cared_at"`
	Plant     Plant     `gorm:"foreignKey:PlantID" json:"plant,omitempty"`
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
