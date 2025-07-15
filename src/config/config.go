package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// DatabaseConfig holds database connection information
type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// AWSSecretFormat represents the structure of database credentials in AWS Secrets Manager
type AWSSecretFormat struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// GetDatabaseConfig retrieves database configuration from either AWS Secrets Manager or environment variables
func GetDatabaseConfig() (*DatabaseConfig, error) {
	useLocalDatabase := os.Getenv("USE_LOCAL_DATABASE")

	if useLocalDatabase == "true" {
		return getLocalDatabaseConfig()
	}

	return getAWSDatabaseConfig()
}

// getLocalDatabaseConfig reads database configuration from environment variables
func getLocalDatabaseConfig() (*DatabaseConfig, error) {
	host := os.Getenv("DB_HOST")
	portStr := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_DATABASE")

	if host == "" || portStr == "" || username == "" || password == "" || database == "" {
		return nil, fmt.Errorf("missing required local database environment variables")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	return &DatabaseConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
	}, nil
}

// getAWSDatabaseConfig retrieves database configuration from AWS Secrets Manager
func getAWSDatabaseConfig() (*DatabaseConfig, error) {
	secretArn := os.Getenv("DB_CREDENTIALS")
	database := os.Getenv("DB_DATABASE")
	awsRegion := os.Getenv("AWS_REGION")

	if secretArn == "" || database == "" || awsRegion == "" {
		return nil, fmt.Errorf("missing required AWS environment variables: DB_CREDENTIALS, DB_DATABASE, AWS_REGION")
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	// Retrieve the secret
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve secret from AWS Secrets Manager: %w", err)
	}

	if result.SecretString == nil {
		return nil, fmt.Errorf("secret string is empty")
	}

	// Parse the secret JSON
	var secret AWSSecretFormat
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		return nil, fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	port, err := strconv.Atoi(secret.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port in secret: %w", err)
	}

	return &DatabaseConfig{
		Host:     secret.Host,
		Port:     port,
		Username: secret.Username,
		Password: secret.Password,
		Database: database,
	}, nil
}
