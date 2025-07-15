package services

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gaulatti/signal/src/models"
	"github.com/gaulatti/signal/src/storage"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
	"gorm.io/gorm"
)

// APNSClient holds a cached APNS client and its config
type APNSClient struct {
	Client     *apns2.Client
	Config     *models.APNSConfig
	LastUsed   time.Time
	PrivateKey *ecdsa.PrivateKey
}

// APNSService handles Apple Push Notification Service integration
type APNSService struct {
	s3Service *storage.S3Service
	db        *gorm.DB
	clients   map[string]*APNSClient // tenantID -> client
	mu        sync.RWMutex
}

// NewAPNSService creates a new APNS service instance
func NewAPNSService(s3Service *storage.S3Service, db *gorm.DB) *APNSService {
	return &APNSService{
		s3Service: s3Service,
		db:        db,
		clients:   make(map[string]*APNSClient),
	}
}

// getOrCreateClient gets or creates an APNS client for a tenant
func (s *APNSService) getOrCreateClient(tenantID string) (*APNSClient, error) {
	s.mu.RLock()
	if client, exists := s.clients[tenantID]; exists {
		client.LastUsed = time.Now()
		s.mu.RUnlock()
		return client, nil
	}
	s.mu.RUnlock()

	// Get APNS config from database
	var config models.APNSConfig
	if err := s.db.Where("tenant_id = ? AND active = ?", tenantID, true).First(&config).Error; err != nil {
		return nil, fmt.Errorf("APNS config not found for tenant %s: %w", tenantID, err)
	}

	// Download .p8 file from S3
	p8Key := fmt.Sprintf("apns/%s.p8", tenantID)
	localPath := filepath.Join("/tmp", "apns", fmt.Sprintf("%s.p8", tenantID))

	if err := s.s3Service.DownloadFile(p8Key, localPath); err != nil {
		return nil, fmt.Errorf("failed to download APNS key for tenant %s: %w", tenantID, err)
	}

	// Load the private key
	authKey, err := token.AuthKeyFromFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load APNS auth key: %w", err)
	}

	// Create token source
	tokenSource := &token.Token{
		AuthKey: authKey,
		KeyID:   config.KeyID,
		TeamID:  config.TeamID,
	}

	// Create APNS client
	client := apns2.NewTokenClient(tokenSource)
	if config.Environment == "sandbox" {
		client = client.Development()
	} else {
		client = client.Production()
	}

	apnsClient := &APNSClient{
		Client:     client,
		Config:     &config,
		LastUsed:   time.Now(),
		PrivateKey: authKey,
	}

	// Cache the client
	s.mu.Lock()
	s.clients[tenantID] = apnsClient
	s.mu.Unlock()

	// Clean up the temporary file
	os.Remove(localPath)

	log.Printf("Created APNS client for tenant %s (env: %s)", tenantID, config.Environment)
	return apnsClient, nil
}

// SendPush sends a push notification via APNS
func (s *APNSService) SendPush(tenantID, deviceToken, title, body string, data map[string]interface{}) error {
	client, err := s.getOrCreateClient(tenantID)
	if err != nil {
		return err
	}

	// Create the notification
	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       client.Config.BundleID,
		Payload: []byte(fmt.Sprintf(`{
			"aps": {
				"alert": {
					"title": "%s",
					"body": "%s"
				},
				"sound": "default"
			}
		}`, title, body)),
	}

	// Add custom data if provided
	if len(data) > 0 {
		// TODO: Properly merge custom data into payload JSON
		log.Printf("Custom data provided but not yet implemented: %+v", data)
	}

	// Send the notification
	res, err := client.Client.Push(notification)
	if err != nil {
		return fmt.Errorf("failed to send APNS push: %w", err)
	}

	if !res.Sent() {
		return fmt.Errorf("APNS push failed: %s (reason: %s)", res.StatusCode, res.Reason)
	}

	log.Printf("✅ APNS push sent successfully to %s via tenant %s", deviceToken, tenantID)
	return nil
}

// CleanupOldClients removes unused clients from cache
func (s *APNSService) CleanupOldClients() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	for tenantID, client := range s.clients {
		if client.LastUsed.Before(cutoff) {
			delete(s.clients, tenantID)
			log.Printf("Cleaned up unused APNS client for tenant %s", tenantID)
		}
	}
}
