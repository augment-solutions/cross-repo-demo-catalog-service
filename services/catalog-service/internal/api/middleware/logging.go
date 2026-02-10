package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Logging returns a middleware that logs HTTP requests
func Logging(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get trace ID if available
		traceID := ""
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().HasTraceID() {
			traceID = span.SpanContext().TraceID().String()
		}

		// Build log fields
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("body_size", c.Writer.Size()),
		}

		if traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// Log based on status code
		status := c.Writer.Status()
		switch {
		case status >= 500:
			logger.Error("server error", fields...)
		case status >= 400:
			logger.Warn("client error", fields...)
		case status >= 300:
			logger.Info("redirect", fields...)
		default:
			logger.Info("request", fields...)
		}
	}
}

// RequestID adds a request ID to the context
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			span := trace.SpanFromContext(c.Request.Context())
			if span.SpanContext().HasTraceID() {
				requestID = span.SpanContext().TraceID().String()
			}
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Recovery returns a middleware that recovers from panics
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				traceID := ""
				span := trace.SpanFromContext(c.Request.Context())
				if span.SpanContext().HasTraceID() {
					traceID = span.SpanContext().TraceID().String()
				}

				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("trace_id", traceID),
				)

				c.AbortWithStatusJSON(500, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}

