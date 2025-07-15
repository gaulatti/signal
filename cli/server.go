package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gaulatti/signal/src/database"
	"github.com/gaulatti/signal/src/handlers"
	"github.com/gaulatti/signal/src/middleware"
	"github.com/gaulatti/signal/src/services"
	"github.com/gaulatti/signal/src/storage"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	} else {
		log.Println("✅ Loaded .env file")
	}

	log.Println("🚀 Starting Signal Push Notification Service...")

	// Initialize database connection
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run database migrations
	log.Println("🔄 Running database migrations...")
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("✅ Database migrations completed")

	// Initialize cache with API keys
	if err := database.InitCache(); err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}

	// Initialize S3 service
	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		s3Bucket = "signal" // default bucket name
	}
	s3Service, err := storage.NewS3Service(s3Bucket)
	if err != nil {
		log.Fatalf("Failed to initialize S3 service: %v", err)
	}
	log.Printf("✅ S3 service initialized (bucket: %s)", s3Bucket)

	// Initialize push notification services
	apnsService := services.NewAPNSService(s3Service, database.DB)
	fcmService := services.NewFCMService(s3Service, database.DB)
	log.Println("✅ Push notification services initialized")

	// Seed tenants from config file if it exists
	seedService := services.NewSeedService(database.DB)
	if err := seedService.SeedTenantsFromFile("./config/tenants.json"); err != nil {
		log.Printf("Warning: Failed to seed tenants: %v", err)
	}

	// Start cleanup goroutine for push service clients
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				apnsService.CleanupOldClients()
				fcmService.CleanupOldClients()
			}
		}
	}()

	// Setup routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	// Protected endpoints that require authentication
	http.HandleFunc("/register", middleware.AuthMiddleware(handlers.RegisterHandler))
	http.HandleFunc("/push", middleware.AuthMiddleware(handlers.PushHandler))

	// New push notification endpoints
	http.HandleFunc("/push/apns", middleware.AuthMiddleware(handlers.APNSPushHandler(apnsService)))
	http.HandleFunc("/push/fcm", middleware.AuthMiddleware(handlers.FCMPushHandler(fcmService)))

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server running on :%s", port)
	log.Printf("📋 Available endpoints:")
	log.Printf("   GET  /health     - Health check (no auth required)")
	log.Printf("   POST /register   - Register device token (auth required)")
	log.Printf("   POST /push       - Send generic push notification (auth required)")
	log.Printf("   POST /push/apns  - Send APNS push notification (auth required)")
	log.Printf("   POST /push/fcm   - Send FCM push notification (auth required)")
	log.Printf("💡 Authentication: Authorization: Digest <md5(api_key + YYYY-MM-DD)>")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
