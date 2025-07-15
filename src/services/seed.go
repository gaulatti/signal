package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gaulatti/signal/src/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TenantSeedData represents the structure of tenant data in local.tenants.json
type TenantSeedData struct {
	TenantID   string          `json:"tenant_id"`
	Name       string          `json:"name"`
	Label      string          `json:"label"`
	APIKey     string          `json:"api_key,omitempty"`
	APNSConfig *APNSConfigSeed `json:"apns_config,omitempty"`
	FCMConfig  *FCMConfigSeed  `json:"fcm_config,omitempty"`
}

type APNSConfigSeed struct {
	TeamID      string `json:"team_id"`
	KeyID       string `json:"key_id"`
	BundleID    string `json:"bundle_id"`
	Environment string `json:"environment"`
}

type FCMConfigSeed struct {
	ProjectID string `json:"project_id"`
	Enabled   bool   `json:"enabled"`
}

// SeedService handles seeding of initial data
type SeedService struct {
	db *gorm.DB
}

// NewSeedService creates a new seed service instance
func NewSeedService(db *gorm.DB) *SeedService {
	return &SeedService{db: db}
}

// SeedTenantsFromFile seeds tenants from a JSON file
func (s *SeedService) SeedTenantsFromFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Seed file %s not found, skipping seeding", filePath)
		return nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	// Parse JSON
	var tenants []TenantSeedData
	if err := json.Unmarshal(data, &tenants); err != nil {
		return fmt.Errorf("failed to parse seed file JSON: %w", err)
	}

	log.Printf("🌱 Seeding %d tenants from %s", len(tenants), filePath)

	// Seed each tenant
	for _, tenantData := range tenants {
		if err := s.seedTenant(tenantData); err != nil {
			log.Printf("Error seeding tenant %s: %v", tenantData.TenantID, err)
			continue
		}
		log.Printf("✅ Seeded tenant: %s (%s)", tenantData.TenantID, tenantData.Name)
	}

	return nil
}

// seedTenant seeds a single tenant with all its configurations
func (s *SeedService) seedTenant(data TenantSeedData) error {
	// Create or update tenant
	tenant := models.Tenant{
		TenantID:    data.TenantID,
		Name:        data.Name,
		Description: fmt.Sprintf("Seeded tenant: %s", data.Name),
		Active:      true,
	}

	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description", "updated_at"}),
	}).Create(&tenant).Error; err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create or update API key if provided
	if data.APIKey != "" {
		apiKey := models.APIKey{
			TenantID: data.TenantID,
			Label:    data.Label,
			APIKey:   data.APIKey,
			Disabled: false,
		}

		if err := s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "api_key"}},
			DoUpdates: clause.AssignmentColumns([]string{"label", "disabled"}),
		}).Create(&apiKey).Error; err != nil {
			return fmt.Errorf("failed to create API key: %w", err)
		}
	}

	// Create or update APNS config if provided
	if data.APNSConfig != nil {
		apnsConfig := models.APNSConfig{
			TenantID:    data.TenantID,
			TeamID:      data.APNSConfig.TeamID,
			KeyID:       data.APNSConfig.KeyID,
			BundleID:    data.APNSConfig.BundleID,
			Environment: data.APNSConfig.Environment,
			Active:      true,
		}

		if err := s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tenant_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"team_id", "key_id", "bundle_id", "environment", "updated_at"}),
		}).Create(&apnsConfig).Error; err != nil {
			return fmt.Errorf("failed to create APNS config: %w", err)
		}
	}

	// Create or update FCM config if provided
	if data.FCMConfig != nil {
		fcmConfig := models.FCMConfig{
			TenantID:  data.TenantID,
			ProjectID: data.FCMConfig.ProjectID,
			Active:    data.FCMConfig.Enabled,
		}

		if err := s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tenant_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"project_id", "active", "updated_at"}),
		}).Create(&fcmConfig).Error; err != nil {
			return fmt.Errorf("failed to create FCM config: %w", err)
		}
	}

	return nil
}
