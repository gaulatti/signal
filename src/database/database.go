package database

import (
	"crypto/md5"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gaulatti/signal/src/config"
	"github.com/gaulatti/signal/src/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Cache holds the in-memory cache of tenant API keys and digest-to-tenant mapping
type Cache struct {
	mu sync.RWMutex
	// tenants: tenantID -> apiKey
	tenants map[string]string
	// digests: digest -> tenantID
	digests map[string]string
}

var (
	DB       *gorm.DB
	APICache *Cache
)

// NewCache creates a new cache instance
func NewCache() *Cache {
	return &Cache{
		tenants: make(map[string]string),
		digests: make(map[string]string),
	}
}

// LoadAPIKeys loads all active API keys into the cache and computes digests
func (c *Cache) LoadAPIKeys(db *gorm.DB) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var apiKeys []models.APIKey
	if err := db.Where("disabled = ?", false).Find(&apiKeys).Error; err != nil {
		return fmt.Errorf("failed to load API keys: %w", err)
	}

	// Clear existing cache
	c.tenants = make(map[string]string)
	c.digests = make(map[string]string)

	currentHour := time.Now().UTC().Format("2006-01-02-15")
	// Also precompute for next hour to allow for clock skew
	nextHour := time.Now().UTC().Add(time.Hour).Format("2006-01-02-15")

	for _, key := range apiKeys {
		c.tenants[key.TenantID] = key.APIKey

		// Compute digest for current hour
		digest := fmt.Sprintf("%x", md5.Sum([]byte(key.APIKey+currentHour)))
		c.digests[digest] = key.TenantID

		// Compute digest for next hour
		digestNext := fmt.Sprintf("%x", md5.Sum([]byte(key.APIKey+nextHour)))
		c.digests[digestNext] = key.TenantID
	}

	log.Printf("Loaded %d active API keys into cache (digests: %d)", len(c.tenants), len(c.digests))
	return nil
}

// GetAPIKey retrieves an API key for a tenant
func (c *Cache) GetAPIKey(tenantID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	apiKey, exists := c.tenants[tenantID]
	return apiKey, exists
}

// GetTenantIDByDigest returns the tenant ID for a given digest, or empty string if not found
func (c *Cache) GetTenantIDByDigest(digest string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.digests[digest]
}

// StartDigestRefresher starts a goroutine to refresh digests every hour
func (c *Cache) StartDigestRefresher(db *gorm.DB) {
	go func() {
		for {
			now := time.Now()
			next := now.Truncate(time.Hour).Add(time.Hour)
			time.Sleep(time.Until(next))
			_ = c.LoadAPIKeys(db)
		}
	}()
}

// GetAllTenants returns all tenant IDs and their API keys (for auth middleware)
func (c *Cache) GetAllTenants() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]string)
	for tenantID, apiKey := range c.tenants {
		result[tenantID] = apiKey
	}
	return result
}

// InitDB initializes the database connection
func InitDB() error {
	dbConfig, err := config.GetDatabaseConfig()
	if err != nil {
		return fmt.Errorf("failed to get database configuration: %w", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("✅ Database connected successfully to %s:%d/%s", dbConfig.Host, dbConfig.Port, dbConfig.Database)
	return nil
}

// InitCache initializes the cache and loads API keys
func InitCache() error {
	APICache = NewCache()
	if err := APICache.LoadAPIKeys(DB); err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}
	APICache.StartDigestRefresher(DB)
	return nil
}

// AutoMigrate runs database migrations
func AutoMigrate() error {
	// Create tables in proper order: parent first, then children
	return DB.AutoMigrate(
		&models.Tenant{},
		&models.APIKey{},
		&models.DeviceToken{},
		&models.APNSConfig{},
		&models.FCMConfig{},
	)
}
