package models

import (
	"time"
)

// APNSConfig represents APNS configuration for tenants
type APNSConfig struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    string    `gorm:"type:varchar(255);not null;index" json:"tenant_id"`
	TeamID      string    `gorm:"type:varchar(255);not null" json:"team_id"`
	KeyID       string    `gorm:"type:varchar(255);not null" json:"key_id"`
	BundleID    string    `gorm:"type:varchar(255);not null" json:"bundle_id"`
	Environment string    `gorm:"type:varchar(100);not null;default:'production'" json:"environment"` // 'production' or 'sandbox'
	Active      bool      `gorm:"default:true" json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
