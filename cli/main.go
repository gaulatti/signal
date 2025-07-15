package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gaulatti/signal/src/database"
	"github.com/gaulatti/signal/src/models"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Parse command line flags
	var (
		tenantID = flag.String("tenant-id", "", "Tenant ID (required)")
		label    = flag.String("label", "", "Label for the API key (required)")
		apiKey   = flag.String("api-key", "", "API key (optional, will generate UUID if not provided)")
		help     = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("Signal API Key Management CLI")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  go run ./cli/main.go -tenant-id=<tenant> -label=<label> [-api-key=<key>]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run ./cli/main.go -tenant-id=product-a -label=\"Production API Key\"")
		fmt.Println("  go run ./cli/main.go -tenant-id=product-b -label=\"Test Key\" -api-key=custom-key-123")
		fmt.Println("")
		fmt.Println("Flags:")
		flag.PrintDefaults()
		return
	}

	if *tenantID == "" || *label == "" {
		fmt.Println("Error: -tenant-id and -label are required")
		fmt.Println("Use -help for usage information")
		os.Exit(1)
	}

	// Generate API key if not provided
	if *apiKey == "" {
		*apiKey = uuid.New().String()
	}

	// Initialize database connection
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations to ensure tables exist
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create or update tenant
	if err := models.CreateTenantIfNotExists(database.DB, *tenantID, *tenantID); err != nil {
		log.Fatalf("Failed to create tenant: %v", err)
	}

	// Create API key
	apiKeyRecord := models.APIKey{
		TenantID: *tenantID,
		Label:    *label,
		APIKey:   *apiKey,
		Disabled: false,
	}

	if err := database.DB.Create(&apiKeyRecord).Error; err != nil {
		log.Fatalf("Failed to create API key: %v", err)
	}

	fmt.Println("✅ API Key created successfully!")
	fmt.Printf("   Tenant ID: %s\n", *tenantID)
	fmt.Printf("   Label:     %s\n", *label)
	fmt.Printf("   API Key:   %s\n", *apiKey)
	fmt.Printf("   ID:        %d\n", apiKeyRecord.ID)
	fmt.Println("")
	fmt.Println("💡 Authentication format:")
	fmt.Println("   Authorization: Digest <md5(api_key + YYYY-MM-DD)>")
}
