package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// GoCardless credentials
	GCSecretID  string
	GCSecretKey string

	// Institution ID to link with
	InstitutionID string

	// HTTP server configuration
	Port int

	// Timeout for waiting for authorization callback
	AuthTimeout time.Duration

	// Timeout for HTTP requests
	HTTPTimeout time.Duration
}

// LoadConfig loads configuration from environment variables and command-line flags
func LoadConfig() (*Config, error) {
	// Load .env file
	err := godotenv.Load("./../.env")
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	// Parse command line arguments
	institutionID := flag.String("institution", "", "Institution ID (required)")
	port := flag.Int("port", 8080, "Port to listen for callback")
	authTimeout := flag.Duration("auth-timeout", 5*time.Minute, "Timeout for waiting for authorization callback")
	httpTimeout := flag.Duration("http-timeout", 20*time.Second, "Timeout for HTTP requests")
	flag.Parse()

	if *institutionID == "" {
		return nil, fmt.Errorf("institution ID is required")
	}

	// Get GoCardless credentials from environment variables
	secretID := os.Getenv("GC_SECRET_ID")
	secretKey := os.Getenv("GC_SECRET_KEY")

	if secretID == "" || secretKey == "" {
		return nil, fmt.Errorf("GC_SECRET_ID and GC_SECRET_KEY environment variables are required")
	}

	return &Config{
		GCSecretID:    secretID,
		GCSecretKey:   secretKey,
		InstitutionID: *institutionID,
		Port:          *port,
		AuthTimeout:   *authTimeout,
		HTTPTimeout:   *httpTimeout,
	}, nil
}
