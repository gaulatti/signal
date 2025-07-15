package models

import (
	"time"
)

// FCMConfig represents FCM configuration for tenants
type FCMConfig struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       string    `gorm:"type:varchar(255);not null;index" json:"tenant_id"`
	ProjectID      string    `gorm:"type:varchar(255);not null" json:"project_id"`
	ServiceAccount string    `gorm:"type:text;not null" json:"service_account"` // JSON content of service account key
	Active         bool      `gorm:"default:true" json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
