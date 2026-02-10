package otel

import (
	"context"
	"net/http"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func TestHTTPHeaderCarrier(t *testing.T) {
	headers := make(http.Header)
	carrier := HTTPHeaderCarrier(headers)

	// Test Set
	carrier.Set("traceparent", "00-abc123-def456-01")
	if got := carrier.Get("traceparent"); got != "00-abc123-def456-01" {
		t.Errorf("Get() = %s, want 00-abc123-def456-01", got)
	}

	// Test Keys
	keys := carrier.Keys()
	if len(keys) != 1 {
		t.Errorf("Keys() returned %d keys, want 1", len(keys))
	}
}

func TestMapCarrier(t *testing.T) {
	carrier := MapCarrier{}

	// Test Set
	carrier.Set("traceparent", "00-abc123-def456-01")
	if got := carrier.Get("traceparent"); got != "00-abc123-def456-01" {
		t.Errorf("Get() = %s, want 00-abc123-def456-01", got)
	}

	// Test Keys
	keys := carrier.Keys()
	if len(keys) != 1 {
		t.Errorf("Keys() returned %d keys, want 1", len(keys))
	}
}

func TestExtractTraceContext(t *testing.T) {
	// Set up propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
	))

	headers := make(http.Header)
	headers.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	ctx := context.Background()
	newCtx := ExtractTraceContext(ctx, headers)

	// Context should be modified (not the same as original)
	if newCtx == nil {
		t.Error("ExtractTraceContext returned nil")
	}
}

func TestInjectTraceContext(t *testing.T) {
	// Set up propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
	))

	headers := make(http.Header)
	ctx := context.Background()

	// Should not panic
	InjectTraceContext(ctx, headers)
}

func TestExtractFromMap(t *testing.T) {
	// Set up propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
	))

	m := map[string]string{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
	}

	ctx := context.Background()
	newCtx := ExtractFromMap(ctx, m)

	if newCtx == nil {
		t.Error("ExtractFromMap returned nil")
	}
}

func TestInjectToMap(t *testing.T) {
	// Set up propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
	))

	m := make(map[string]string)
	ctx := context.Background()

	// Should not panic
	InjectToMap(ctx, m)
}

func TestSpanFromContext(t *testing.T) {
	ctx := context.Background()
	span := SpanFromContext(ctx)

	// Should return a non-recording span from empty context
	if span == nil {
		t.Error("SpanFromContext returned nil")
	}
}

func TestAddSpanEvent(t *testing.T) {
	ctx := context.Background()
	// Should not panic even with no active span
	AddSpanEvent(ctx, "test-event")
}

func TestSetSpanError(t *testing.T) {
	ctx := context.Background()
	err := context.DeadlineExceeded
	// Should not panic even with no active span
	SetSpanError(ctx, err)
}

