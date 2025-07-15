package models

import (
	"time"
)

// APIKey represents the api_keys table
type APIKey struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  string    `gorm:"type:varchar(255);not null;index" json:"tenant_id"`
	Label     string    `gorm:"type:varchar(500)" json:"label"`
	APIKey    string    `gorm:"type:varchar(500);not null;uniqueIndex" json:"api_key"`
	Disabled  bool      `gorm:"default:false" json:"disabled"`
	CreatedAt time.Time `json:"created_at"`
}
