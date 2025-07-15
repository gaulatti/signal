package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gaulatti/signal/src/database"
	"github.com/gaulatti/signal/src/middleware"
	"github.com/gaulatti/signal/src/models"
)

// PushRequest represents the push notification payload
type PushRequest struct {
	UserID string                 `json:"user_id,omitempty"`
	Title  string                 `json:"title"`
	Body   string                 `json:"body"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// PushHandler handles push notification requests
func PushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		http.Error(w, "Unable to determine tenant", http.StatusInternalServerError)
		return
	}

	var req PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Body == "" {
		http.Error(w, "Missing required fields: title, body", http.StatusBadRequest)
		return
	}

	// Find target devices
	var devices []models.DeviceToken
	query := database.DB.Where("tenant_id = ?", tenantID)

	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	if err := query.Find(&devices).Error; err != nil {
		log.Printf("Error finding devices: %v", err)
		http.Error(w, "Failed to find target devices", http.StatusInternalServerError)
		return
	}

	if len(devices) == 0 {
		http.Error(w, "No devices found for push notification", http.StatusNotFound)
		return
	}

	// Simulate push notification (just log for now)
	for _, device := range devices {
		log.Printf("📱 [SIMULATED PUSH] Tenant: %s, User: %s, Platform: %s, Device: %s, Title: %s, Body: %s, Data: %+v",
			tenantID, device.UserID, device.Platform, device.DeviceToken, req.Title, req.Body, req.Data)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "Push notification sent",
		"devices_sent": len(devices),
	})
}
