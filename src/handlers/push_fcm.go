package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gaulatti/signal/src/middleware"
	"github.com/gaulatti/signal/src/services"
)

// FCMPushRequest represents the FCM push notification payload
type FCMPushRequest struct {
	UserID      string                 `json:"user_id"`
	DeviceToken string                 `json:"device_token"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// FCMPushHandler handles FCM push notification requests
func FCMPushHandler(fcmService *services.FCMService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tenantID := middleware.GetTenantID(r)
		if tenantID == "" {
			http.Error(w, "Unable to determine tenant", http.StatusInternalServerError)
			return
		}

		var req FCMPushRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if req.DeviceToken == "" || req.Title == "" || req.Body == "" {
			http.Error(w, "Missing required fields: device_token, title, body", http.StatusBadRequest)
			return
		}

		// Send FCM push notification
		if err := fcmService.SendPush(tenantID, req.DeviceToken, req.Title, req.Body, req.Data); err != nil {
			log.Printf("Error sending FCM push for tenant %s: %v", tenantID, err)
			http.Error(w, "Failed to send push notification", http.StatusInternalServerError)
			return
		}

		log.Printf("FCM push sent for tenant %s: user=%s, device=%s", tenantID, req.UserID, req.DeviceToken)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "FCM push notification sent successfully",
			"tenant":  tenantID,
		})
	}
}
