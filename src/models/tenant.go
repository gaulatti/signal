package models

import (
	"time"

	"gorm.io/gorm"
)

// Tenant represents the tenants table
type Tenant struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"tenant_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Active      bool      `gorm:"default:true" json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTenantIfNotExists creates a tenant if it doesn't exist
func CreateTenantIfNotExists(db *gorm.DB, tenantID, name string) error {
	var tenant Tenant
	result := db.Where("tenant_id = ?", tenantID).First(&tenant)

	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		// Tenant doesn't exist, create it
		tenant = Tenant{
			TenantID: tenantID,
			Name:     name,
			Active:   true,
		}
		return db.Create(&tenant).Error
	}

	return result.Error
}

// GetActiveTenants returns all active tenants
func GetActiveTenants(db *gorm.DB) ([]Tenant, error) {
	var tenants []Tenant
	err := db.Where("active = ?", true).Find(&tenants).Error
	return tenants, err
}

// GetTenantByID returns a tenant by tenant_id
func GetTenantByID(db *gorm.DB, tenantID string) (*Tenant, error) {
	var tenant Tenant
	err := db.Where("tenant_id = ?", tenantID).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}
