package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port    string
	GinMode string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// MinIO / S3
	StorageType    string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string

	// JWT
	JWTSecret        string
	JWTExpiry        string
	JWTRefreshExpiry string

	// Scanner Service
	ScannerServiceURL string

	// Scan Configuration
	ScanSandboxTimeout   int
	ScanMaxConcurrent    int
	AutoApproveMaxScore  int

	// OIDC
	OIDCEnabled  bool
	OIDCIssuer   string
	OIDCClientID string
	OIDCSecret   string

	// Notification
	SMTPHost            string
	SMTPPort            string
	SMTPUser            string
	SMTPPassword        string
	NotificationEmailFrom string
	SlackWebhookURL     string

	// GitHub
	GitHubToken string
}

func Load() (*Config, error) {
	// Try to load .env file (ignore error if not exists)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                getEnv("PORT", "3000"),
		GinMode:             getEnv("GIN_MODE", "debug"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgresql://skillshub:skillshub_password@localhost:5432/skillshub"),
		RedisURL:            getEnv("REDIS_URL", "redis://localhost:6379"),
		StorageType:         getEnv("STORAGE_TYPE", "minio"),
		MinIOEndpoint:       getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:      getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:      getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:         getEnv("MINIO_BUCKET", "skillshub-files"),
		JWTSecret:           getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTExpiry:           getEnv("JWT_EXPIRY", "8h"),
		JWTRefreshExpiry:    getEnv("JWT_REFRESH_EXPIRY", "7d"),
		ScannerServiceURL:   getEnv("SCANNER_SERVICE_URL", "http://localhost:8000"),
		ScanSandboxTimeout:  getEnvInt("SCAN_SANDBOX_TIMEOUT", 120000),
		ScanMaxConcurrent:   getEnvInt("SCAN_MAX_CONCURRENT", 5),
		AutoApproveMaxScore: getEnvInt("AUTO_APPROVE_MAX_SCORE", 30),
		OIDCEnabled:         getEnvBool("OIDC_ENABLED", false),
		OIDCIssuer:          getEnv("OIDC_ISSUER", ""),
		OIDCClientID:        getEnv("OIDC_CLIENT_ID", ""),
		OIDCSecret:          getEnv("OIDC_CLIENT_SECRET", ""),
		SMTPHost:            getEnv("SMTP_HOST", ""),
		SMTPPort:            getEnv("SMTP_PORT", "587"),
		SMTPUser:            getEnv("SMTP_USER", ""),
		SMTPPassword:        getEnv("SMTP_PASSWORD", ""),
		NotificationEmailFrom: getEnv("NOTIFICATION_EMAIL_FROM", ""),
		SlackWebhookURL:     getEnv("SLACK_WEBHOOK_URL", ""),
		GitHubToken:         getEnv("GITHUB_TOKEN", ""),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result int
	fmt.Sscanf(value, "%d", &result)
	return result
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1"
}
