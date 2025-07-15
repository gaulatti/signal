package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/gaulatti/signal/src/models"
	"github.com/gaulatti/signal/src/storage"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

// FCMClient holds a cached FCM client and its config
type FCMClient struct {
	Client   *messaging.Client
	Config   *models.FCMConfig
	LastUsed time.Time
}

// FCMService handles Firebase Cloud Messaging integration
type FCMService struct {
	s3Service *storage.S3Service
	db        *gorm.DB
	clients   map[string]*FCMClient // tenantID -> client
	mu        sync.RWMutex
}

// NewFCMService creates a new FCM service instance
func NewFCMService(s3Service *storage.S3Service, db *gorm.DB) *FCMService {
	return &FCMService{
		s3Service: s3Service,
		db:        db,
		clients:   make(map[string]*FCMClient),
	}
}

// getOrCreateClient gets or creates an FCM client for a tenant
func (s *FCMService) getOrCreateClient(tenantID string) (*FCMClient, error) {
	s.mu.RLock()
	if client, exists := s.clients[tenantID]; exists {
		client.LastUsed = time.Now()
		s.mu.RUnlock()
		return client, nil
	}
	s.mu.RUnlock()

	// Get FCM config from database
	var config models.FCMConfig
	if err := s.db.Where("tenant_id = ? AND active = ?", tenantID, true).First(&config).Error; err != nil {
		return nil, fmt.Errorf("FCM config not found for tenant %s: %w", tenantID, err)
	}

	// Download service account JSON from S3
	jsonKey := fmt.Sprintf("fcm/%s.json", tenantID)
	serviceAccountJSON, err := s.s3Service.GetFileContent(jsonKey)
	if err != nil {
		return nil, fmt.Errorf("failed to download FCM service account for tenant %s: %w", tenantID, err)
	}

	// Initialize Firebase app
	opt := option.WithCredentialsJSON(serviceAccountJSON)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	// Get messaging client
	ctx := context.Background()
	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get FCM messaging client: %w", err)
	}

	fcmClient := &FCMClient{
		Client:   messagingClient,
		Config:   &config,
		LastUsed: time.Now(),
	}

	// Cache the client
	s.mu.Lock()
	s.clients[tenantID] = fcmClient
	s.mu.Unlock()

	log.Printf("Created FCM client for tenant %s (project: %s)", tenantID, config.ProjectID)
	return fcmClient, nil
}

// SendPush sends a push notification via FCM
func (s *FCMService) SendPush(tenantID, deviceToken, title, body string, data map[string]interface{}) error {
	client, err := s.getOrCreateClient(tenantID)
	if err != nil {
		return err
	}

	// Convert data map to string map (FCM requirement)
	stringData := make(map[string]string)
	for k, v := range data {
		stringData[k] = fmt.Sprintf("%v", v)
	}

	// Create the message
	message := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: stringData,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	// Send the message
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.Client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send FCM push: %w", err)
	}

	log.Printf("✅ FCM push sent successfully to %s via tenant %s (response: %s)", deviceToken, tenantID, response)
	return nil
}

// CleanupOldClients removes unused clients from cache
func (s *FCMService) CleanupOldClients() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	for tenantID, client := range s.clients {
		if client.LastUsed.Before(cutoff) {
			delete(s.clients, tenantID)
			log.Printf("Cleaned up unused FCM client for tenant %s", tenantID)
		}
	}
}
