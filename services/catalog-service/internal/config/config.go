package config

import "os"

// Config holds all configuration for the catalog service
type Config struct {
	ServiceName      string
	ServiceVersion   string
	Environment      string
	Host             string
	Port             string
	OTLPEndpoint     string
	DatabaseURL      string
	RedisURL         string
	ElasticsearchURL string
	LogLevel         string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		ServiceName:      getEnvOrDefault("SERVICE_NAME", "catalog-service"),
		ServiceVersion:   getEnvOrDefault("SERVICE_VERSION", "1.0.0"),
		Environment:      getEnvOrDefault("ENVIRONMENT", "development"),
		Host:             getEnvOrDefault("HOST", "0.0.0.0"),
		Port:             getEnvOrDefault("PORT", "8080"),
		OTLPEndpoint:     getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"),
		DatabaseURL:      getEnvOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/catalog?sslmode=disable"),
		RedisURL:         getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
		ElasticsearchURL: getEnvOrDefault("ELASTICSEARCH_URL", "http://localhost:9200"),
		LogLevel:         getEnvOrDefault("LOG_LEVEL", "info"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

