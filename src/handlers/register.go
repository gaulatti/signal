package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gaulatti/signal/src/database"
	"github.com/gaulatti/signal/src/middleware"
	"github.com/gaulatti/signal/src/models"
)

// RegisterRequest represents the device registration payload
type RegisterRequest struct {
	DeviceToken string `json:"device_token"`
	UserID      string `json:"user_id"`
	Platform    string `json:"platform"`
}

// RegisterHandler handles device token registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		http.Error(w, "Unable to determine tenant", http.StatusInternalServerError)
		return
	}

	// Verify tenant exists and is active
	tenant, err := models.GetTenantByID(database.DB, tenantID)
	if err != nil {
		log.Printf("Error finding tenant %s: %v", tenantID, err)
		http.Error(w, "Invalid tenant", http.StatusUnauthorized)
		return
	}

	if !tenant.Active {
		http.Error(w, "Tenant is not active", http.StatusForbidden)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.DeviceToken == "" || req.UserID == "" || req.Platform == "" {
		http.Error(w, "Missing required fields: device_token, user_id, platform", http.StatusBadRequest)
		return
	}

	// Create or update device token
	deviceToken := models.DeviceToken{
		TenantID:    tenantID,
		DeviceToken: req.DeviceToken,
		UserID:      req.UserID,
		Platform:    req.Platform,
	}

	// Use GORM's upsert functionality
	result := database.DB.Where("tenant_id = ? AND user_id = ? AND platform = ?",
		tenantID, req.UserID, req.Platform).Assign(deviceToken).FirstOrCreate(&deviceToken)

	if result.Error != nil {
		log.Printf("Error saving device token: %v", result.Error)
		http.Error(w, "Failed to register device", http.StatusInternalServerError)
		return
	}

	log.Printf("Device registered for tenant %s: user=%s, platform=%s", tenantID, req.UserID, req.Platform)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Device registered successfully",
		"id":      deviceToken.ID,
		"tenant":  tenant.Name,
	})
}
