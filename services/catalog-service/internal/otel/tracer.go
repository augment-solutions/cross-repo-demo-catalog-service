// Package otel provides OpenTelemetry instrumentation for Catalog Service.
package otel

import (
	"context"
	"net/url"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer         trace.Tracer
	meter          metric.Meter
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
)

// Config holds OpenTelemetry configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
}

// IsDisabled returns true if OpenTelemetry SDK is disabled.
func IsDisabled() bool {
	disabled := os.Getenv("OTEL_SDK_DISABLED")
	return disabled == "true" || disabled == "1"
}

// InitTracer initializes OpenTelemetry tracer and meter providers.
func InitTracer(ctx context.Context, cfg Config) error {
	if IsDisabled() {
		// SDK is disabled, set up no-op providers
		tracer = otel.Tracer(cfg.ServiceName)
		meter = otel.Meter(cfg.ServiceName)
		return nil
	}

	if cfg.OTLPEndpoint == "" {
		cfg.OTLPEndpoint = getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = getEnvOrDefault("OTEL_SERVICE_NAME", "catalog-service")
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("deployment.environment", cfg.Environment),
			attribute.String("service.namespace", "ecomm-stack"),
		),
	)
	if err != nil {
		return err
	}

	endpoint := parseEndpoint(cfg.OTLPEndpoint)
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return err
	}

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return err
	}

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(60*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	tracer = otel.Tracer(cfg.ServiceName)
	meter = otel.Meter(cfg.ServiceName)

	if err := initMetrics(); err != nil {
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the tracer and meter providers.
func Shutdown(ctx context.Context) error {
	if tracerProvider != nil {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if meterProvider != nil {
		if err := meterProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}

func parseEndpoint(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		if u, err := url.Parse(endpoint); err == nil {
			return u.Host
		}
	}
	return endpoint
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetTracer returns the global tracer.
func GetTracer() trace.Tracer {
	return tracer
}

// GetMeter returns the global meter.
func GetMeter() metric.Meter {
	return meter
}

