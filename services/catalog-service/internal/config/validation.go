package config

import (
	"errors"
	"net/url"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errs []string

	if c.ServiceName == "" {
		errs = append(errs, "SERVICE_NAME is required")
	}

	if c.Port == "" {
		errs = append(errs, "PORT is required")
	}

	if c.DatabaseURL == "" {
		errs = append(errs, "DATABASE_URL is required")
	} else if _, err := url.Parse(c.DatabaseURL); err != nil {
		errs = append(errs, "DATABASE_URL is not a valid URL")
	}

	if c.RedisURL == "" {
		errs = append(errs, "REDIS_URL is required")
	}

	if c.OTLPEndpoint == "" {
		errs = append(errs, "OTEL_EXPORTER_OTLP_ENDPOINT is required")
	}

	validEnvironments := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvironments[c.Environment] {
		errs = append(errs, "ENVIRONMENT must be one of: development, staging, production")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

