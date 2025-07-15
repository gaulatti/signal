package services

import (
	"github.com/gaulatti/signal/src/database"
	"github.com/gaulatti/signal/src/models"
)

// TenantLoader handles tenant management operations
type TenantLoader struct{}

// NewTenantLoader creates a new tenant loader instance
func NewTenantLoader() *TenantLoader {
	return &TenantLoader{}
}

// LoadTenant loads or creates a tenant
func (tl *TenantLoader) LoadTenant(tenantID, name string) error {
	return models.CreateTenantIfNotExists(database.DB, tenantID, name)
}

// GetActiveTenants returns all active tenants
func (tl *TenantLoader) GetActiveTenants() ([]models.Tenant, error) {
	return models.GetActiveTenants(database.DB)
}

// GetTenant returns a tenant by ID
func (tl *TenantLoader) GetTenant(tenantID string) (*models.Tenant, error) {
	return models.GetTenantByID(database.DB, tenantID)
}
