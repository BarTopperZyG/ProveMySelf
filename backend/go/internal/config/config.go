package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Application
	Environment string
	Port        string
	LogLevel    string

	// Database
	DatabaseURL string

	// Storage
	StorageType string
	StoragePath string
	S3Bucket    string
	S3Region    string

	// xAPI
	LRSEndpoint  string
	LRSAuthToken string

	// Real-time Collaboration
	YjsProviderURL string

	// Security
	JWTSecret   string
	CORSOrigins []string

	// Email
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string

	// Feature Flags
	EnableCollaboration  bool
	EnableAnalytics      bool
	EnableLTIIntegration bool

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   int

	// File Upload
	MaxFileSize      int64
	AllowedFileTypes []string
}

func Load() (*Config, error) {
	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		StorageType: getEnv("STORAGE_TYPE", "local"),
		StoragePath: getEnv("STORAGE_PATH", "./storage"),
		S3Bucket:    getEnv("S3_BUCKET", ""),
		S3Region:    getEnv("S3_REGION", ""),

		LRSEndpoint:  getEnv("LRS_ENDPOINT", ""),
		LRSAuthToken: getEnv("LRS_AUTH_TOKEN", ""),

		YjsProviderURL: getEnv("YLOG_PROVIDER_URL", ""),

		JWTSecret:   getEnv("JWT_SECRET", ""),
		CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:3001"), ","),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@provemyself.com"),

		EnableCollaboration:  getEnvBool("ENABLE_COLLABORATION", true),
		EnableAnalytics:      getEnvBool("ENABLE_ANALYTICS", true),
		EnableLTIIntegration: getEnvBool("ENABLE_LTI_INTEGRATION", false),

		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvInt("RATE_LIMIT_WINDOW", 60),

		MaxFileSize:      int64(getEnvInt("MAX_FILE_SIZE", 10485760)), // 10MB default
		AllowedFileTypes: strings.Split(getEnv("ALLOWED_FILE_TYPES", "image/jpeg,image/png,image/gif,image/webp"), ","),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that required configuration values are present
func (c *Config) Validate() error {
	if c.Environment == "production" {
		if c.JWTSecret == "" {
			return errors.New("JWT_SECRET is required in production")
		}
		if len(c.JWTSecret) < 32 {
			return errors.New("JWT_SECRET must be at least 32 characters in production")
		}
		if c.DatabaseURL == "" {
			return errors.New("DATABASE_URL is required in production")
		}
	}

	if c.StorageType == "s3" {
		if c.S3Bucket == "" {
			return errors.New("S3_BUCKET is required when STORAGE_TYPE=s3")
		}
		if c.S3Region == "" {
			return errors.New("S3_REGION is required when STORAGE_TYPE=s3")
		}
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsTest returns true if running in test mode
func (c *Config) IsTest() bool {
	return c.Environment == "test"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}