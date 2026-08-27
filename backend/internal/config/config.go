package config

import (
	"log"
	"os"

	"coherence.dev/backend/internal/providers"
)

// Config holds application configuration
type Config struct {
	DatabaseURL string
	RedisURL    string
	Environment string
	Port        string
	LogLevel    string
	
	AWSConfig   providers.AWSConfig
	GCPConfig   providers.GCPConfig
	AzureConfig providers.AzureConfig
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://coherence:coherence@localhost:5432/coherence"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		
		AWSConfig: providers.AWSConfig{
			Region:      getEnv("AWS_REGION", "us-east-1"),
			Profile:     getEnv("AWS_PROFILE", "default"),
		},
		GCPConfig: providers.GCPConfig{
			ProjectID:   getEnv("GCP_PROJECT_ID", ""),
			Credentials: getEnv("GCP_CREDENTIALS", ""),
		},
		AzureConfig: providers.AzureConfig{
			SubscriptionID: getEnv("AZURE_SUBSCRIPTION_ID", ""),
			TenantID:       getEnv("AZURE_TENANT_ID", ""),
			ClientID:       getEnv("AZURE_CLIENT_ID", ""),
			ClientSecret:   getEnv("AZURE_CLIENT_SECRET", ""),
		},
	}
	
	log.Printf("Configuration loaded: env=%s, port=%s", cfg.Environment, cfg.Port)
	return cfg
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
