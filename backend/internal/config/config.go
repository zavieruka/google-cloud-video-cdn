package config

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
)

// Config holds all application configuration
type Config struct {
	// GCP Configuration
	GCPProjectID        string
	GCPRegion           string
	FirestoreDatabaseID string
	SourceBucketName    string
	ProcessedBucketName string

	// Application Configuration
	Port        string
	Environment string
	LogLevel    string

	// GCP Clients (initialized after validation)
	FirestoreClient *firestore.Client
	StorageClient   *storage.Client
}

// Load reads configuration from environment variables and validates them
func Load() (*Config, error) {
	cfg := &Config{
		GCPProjectID:        getEnv("GCP_PROJECT_ID", ""),
		GCPRegion:           getEnv("GCP_REGION", "us-central1"),
		FirestoreDatabaseID: getEnv("FIRESTORE_DATABASE_ID", "(default)"),
		SourceBucketName:    getEnv("SOURCE_BUCKET_NAME", ""),
		ProcessedBucketName: getEnv("PROCESSED_BUCKET_NAME", ""),
		Port:                getEnv("PORT", "8080"),
		Environment:         getEnv("ENVIRONMENT", "dev"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required configuration is present and valid
func (c *Config) Validate() error {
	if c.GCPProjectID == "" {
		return fmt.Errorf("GCP_PROJECT_ID is required")
	}

	if c.SourceBucketName == "" {
		return fmt.Errorf("SOURCE_BUCKET_NAME is required")
	}

	if c.ProcessedBucketName == "" {
		return fmt.Errorf("PROCESSED_BUCKET_NAME is required")
	}

	// Validate port is a valid number
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("PORT must be a valid number: %w", err)
	}

	// Validate environment
	validEnvs := map[string]bool{"dev": true, "staging": true, "production": true}
	if !validEnvs[c.Environment] {
		return fmt.Errorf("ENVIRONMENT must be one of: dev, staging, production (got: %s)", c.Environment)
	}

	// Validate log level
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error (got: %s)", c.LogLevel)
	}

	return nil
}

// InitializeGCPClients creates and initializes GCP service clients
// This should be called after Load() and Validate()
func (c *Config) InitializeGCPClients(ctx context.Context) error {
	var err error

	// Initialize Firestore client with specified database
	c.FirestoreClient, err = firestore.NewClientWithDatabase(ctx, c.GCPProjectID, c.FirestoreDatabaseID)
	if err != nil {
		return fmt.Errorf("failed to create Firestore client for database '%s': %w", c.FirestoreDatabaseID, err)
	}

	// Initialize Cloud Storage client
	c.StorageClient, err = storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Storage client: %w", err)
	}

	return nil
}

// Close gracefully closes all GCP client connections
func (c *Config) Close() error {
	var errs []error

	if c.FirestoreClient != nil {
		if err := c.FirestoreClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Firestore client: %w", err))
		}
	}

	if c.StorageClient != nil {
		if err := c.StorageClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Storage client: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "dev"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// GetAddress returns the full address the server should listen on
func (c *Config) GetAddress() string {
	return fmt.Sprintf(":%s", c.Port)
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
