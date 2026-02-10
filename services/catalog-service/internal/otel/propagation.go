package otel

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HTTPHeaderCarrier adapts http.Header for trace context propagation
type HTTPHeaderCarrier http.Header

// Get returns the value associated with the passed key.
func (c HTTPHeaderCarrier) Get(key string) string {
	return http.Header(c).Get(key)
}

// Set stores the key-value pair.
func (c HTTPHeaderCarrier) Set(key, value string) {
	http.Header(c).Set(key, value)
}

// Keys returns a slice of keys in the carrier.
func (c HTTPHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// ExtractTraceContext extracts trace context from HTTP headers
func ExtractTraceContext(ctx context.Context, headers http.Header) context.Context {
	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, HTTPHeaderCarrier(headers))
}

// InjectTraceContext injects trace context into HTTP headers
func InjectTraceContext(ctx context.Context, headers http.Header) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, HTTPHeaderCarrier(headers))
}

// SpanFromContext returns the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, opts...)
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(ctx context.Context, name string, attrs ...trace.EventOption) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, attrs...)
}

// SetSpanError marks the span as having an error
func SetSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

// MapCarrier is a simple map-based carrier for propagation
type MapCarrier map[string]string

// Get returns the value for a key
func (c MapCarrier) Get(key string) string {
	return c[key]
}

// Set sets a key-value pair
func (c MapCarrier) Set(key, value string) {
	c[key] = value
}

// Keys returns all keys
func (c MapCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// ExtractFromMap extracts trace context from a map
func ExtractFromMap(ctx context.Context, m map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(m))
}

// InjectToMap injects trace context into a map
func InjectToMap(ctx context.Context, m map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(m))
}

