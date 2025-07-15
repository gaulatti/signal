package models

import (
	"time"
)

// DeviceToken represents device registrations scoped to tenants
type DeviceToken struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    string    `gorm:"type:varchar(255);not null;index" json:"tenant_id"`
	DeviceToken string    `gorm:"type:varchar(500);not null" json:"device_token"`
	UserID      string    `gorm:"type:varchar(255);not null" json:"user_id"`
	Platform    string    `gorm:"type:varchar(100);not null" json:"platform"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
