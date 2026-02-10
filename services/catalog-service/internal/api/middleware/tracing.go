package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracing returns a middleware that creates spans for HTTP requests
func Tracing(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// Extract context from incoming request headers
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Create span name from method and path
		spanName := c.Request.Method + " " + c.FullPath()
		if spanName == "" {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		}

		// Start span with HTTP attributes
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPURL(c.Request.URL.String()),
				semconv.HTTPScheme(c.Request.URL.Scheme),
				attribute.String("http.host", c.Request.Host),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Set response attributes
		status := c.Writer.Status()
		span.SetAttributes(semconv.HTTPStatusCode(status))

		// Record errors
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("error.message", c.Errors.String()))
			span.RecordError(c.Errors.Last().Err)
		}

		// Set span status based on HTTP status code
		if status >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}

// InjectTraceID injects trace ID into response headers
func InjectTraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().HasTraceID() {
			c.Header("X-Trace-ID", span.SpanContext().TraceID().String())
		}
		c.Next()
	}
}

// ExtractTraceContext extracts trace context from request headers
func ExtractTraceContext() gin.HandlerFunc {
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

