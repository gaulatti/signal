package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gaulatti/signal/src/database"
)

type contextKey string

const tenantIDKey contextKey = "tenant_id"

// AuthMiddleware handles API key authentication
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Parse "Digest <digest>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Digest" {
			http.Error(w, "Invalid Authorization header format. Expected: Digest <digest>", http.StatusUnauthorized)
			return
		}

		providedDigest := parts[1]

		// O(1) lookup in digest cache
		tenantID := database.APICache.GetTenantIDByDigest(providedDigest)
		if tenantID == "" {
			http.Error(w, "Unauthorized: invalid or expired API key", http.StatusUnauthorized)
			return
		}

		// Set tenant ID in context
		ctx := context.WithValue(r.Context(), tenantIDKey, tenantID)
		next(w, r.WithContext(ctx))
	}
}

// GetTenantID extracts tenant ID from request context
func GetTenantID(r *http.Request) string {
	if tenantID, ok := r.Context().Value(tenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}
