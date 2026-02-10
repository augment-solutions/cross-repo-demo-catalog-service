package otel

import (
	"context"
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Config{}

	if cfg.ServiceName != "" {
		t.Errorf("expected empty ServiceName, got %s", cfg.ServiceName)
	}
	if cfg.OTLPEndpoint != "" {
		t.Errorf("expected empty OTLPEndpoint, got %s", cfg.OTLPEndpoint)
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "returns env value when set",
			key:          "TEST_CATALOG_OTEL_KEY",
			envValue:     "custom-value",
			defaultValue: "default",
			expected:     "custom-value",
		},
		{
			name:         "returns default when env not set",
			key:          "TEST_CATALOG_OTEL_UNSET_KEY",
			envValue:     "",
			defaultValue: "default-value",
			expected:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}
			result := getEnvOrDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "host:port format unchanged",
			input:    "localhost:4317",
			expected: "localhost:4317",
		},
		{
			name:     "http URL extracts host:port",
			input:    "http://otel-collector:4317",
			expected: "otel-collector:4317",
		},
		{
			name:     "https URL extracts host:port",
			input:    "https://otel.example.com:443",
			expected: "otel.example.com:443",
		},
		{
			name:     "http URL with path extracts host:port",
			input:    "http://localhost:4317/v1/traces",
			expected: "localhost:4317",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEndpoint(tt.input)
			if result != tt.expected {
				t.Errorf("parseEndpoint(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetTracer(t *testing.T) {
	tr := GetTracer()
	if tr != nil {
		t.Log("Tracer is set - likely from previous test or global state")
	}
}

func TestGetMeter(t *testing.T) {
	m := GetMeter()
	if m != nil {
		t.Log("Meter is set - likely from previous test or global state")
	}
}

func TestShutdown(t *testing.T) {
	ctx := context.Background()
	err := Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown returned unexpected error: %v", err)
	}
}

func TestConfigFields(t *testing.T) {
	cfg := Config{
		ServiceName:    "test-catalog-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4317",
	}

	if cfg.ServiceName != "test-catalog-service" {
		t.Errorf("expected ServiceName 'test-catalog-service', got %s", cfg.ServiceName)
	}
	if cfg.ServiceVersion != "1.0.0" {
		t.Errorf("expected ServiceVersion '1.0.0', got %s", cfg.ServiceVersion)
	}
	if cfg.Environment != "test" {
		t.Errorf("expected Environment 'test', got %s", cfg.Environment)
	}
	if cfg.OTLPEndpoint != "localhost:4317" {
		t.Errorf("expected OTLPEndpoint 'localhost:4317', got %s", cfg.OTLPEndpoint)
	}
}

