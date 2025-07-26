package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration parameters for the application
type Config struct {
	// GoCardless configuration
	GCSecretID  string
	GCSecretKey string

	// YNAB configuration
	YNABToken string

	// Synchronization configuration
	CronSchedule string
	Jobs         []job

	// Monitoring configuration
	NewRelicLicenseKey string
	NewRelicAppName    string
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() (Config, error) {
	// Load .env file
	err := godotenv.Load(".env")
	if err != nil {
		return Config{}, fmt.Errorf("failed to load .env file: %w", err)
	}

	secretID := os.Getenv("GC_SECRET_ID")
	secretKey := os.Getenv("GC_SECRET_KEY")
	ynabToken := os.Getenv("YNAB_TOKEN")
	cronSchedule := os.Getenv("CRON_SCHEDULE")
	newRelicLicenseKey := os.Getenv("NEW_RELIC_LICENCE_KEY")
	newRelicAppName := os.Getenv("NEW_RELIC_APP_NAME")

	// Validate required configuration
	if secretID == "" || secretKey == "" || ynabToken == "" {
		return Config{}, fmt.Errorf("GC_SECRET_ID, GC_SECRET_KEY, and YNAB_TOKEN environment variables are required")
	}

	// Parse jobs from the environment
	jobs, err := envToJobs(os.Getenv("JOBS"))
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse jobs: %w", err)
	}

	// Set the default cron schedule if not provided
	if cronSchedule == "" {
		cronSchedule = "0 6,18 * * *"
	}

	return Config{
		GCSecretID:         secretID,
		GCSecretKey:        secretKey,
		YNABToken:          ynabToken,
		CronSchedule:       cronSchedule,
		Jobs:               jobs,
		NewRelicLicenseKey: newRelicLicenseKey,
		NewRelicAppName:    newRelicAppName,
	}, nil
}
