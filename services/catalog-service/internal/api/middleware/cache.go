package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// CacheConfig holds cache middleware configuration
type CacheConfig struct {
	DefaultTTL time.Duration
	Prefix     string
}

// responseWriter wraps gin.ResponseWriter to capture response
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Cache returns a caching middleware
func Cache(client *redis.Client, cfg CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Skip if cache control header is set
		if c.GetHeader("Cache-Control") == "no-cache" {
			c.Next()
			return
		}

		// Generate cache key
		key := generateCacheKey(cfg.Prefix, c.Request)

		// Try to get from cache
		ctx := c.Request.Context()
		cached, err := client.Get(ctx, key).Result()
		if err == nil {
			// Cache hit
			c.Header("X-Cache", "HIT")
			c.Header("Content-Type", "application/json")
			c.String(http.StatusOK, cached)
			c.Abort()
			return
		}

		// Cache miss - capture response
		c.Header("X-Cache", "MISS")
		w := &responseWriter{body: bytes.NewBuffer(nil), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		// Only cache successful responses
		if c.Writer.Status() == http.StatusOK {
			ttl := cfg.DefaultTTL
			if ttlHeader := c.GetHeader("X-Cache-TTL"); ttlHeader != "" {
				if parsed, err := time.ParseDuration(ttlHeader); err == nil {
					ttl = parsed
				}
			}

			// Store in cache
			client.Set(ctx, key, w.body.String(), ttl)
		}
	}
}

// CacheInvalidate returns a middleware that invalidates cache on mutations
func CacheInvalidate(client *redis.Client, prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request first
		c.Next()

		// Invalidate cache on successful mutations
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			if c.Request.Method == http.MethodPost ||
				c.Request.Method == http.MethodPut ||
				c.Request.Method == http.MethodPatch ||
				c.Request.Method == http.MethodDelete {
				// Invalidate related cache entries
				pattern := prefix + "*"
				invalidatePattern(c.Request.Context(), client, pattern)
			}
		}
	}
}

// generateCacheKey generates a unique cache key for a request
func generateCacheKey(prefix string, r *http.Request) string {
	// Include path and query parameters
	data := r.URL.Path + r.URL.RawQuery

	// Hash the key
	hash := sha256.Sum256([]byte(data))
	return prefix + hex.EncodeToString(hash[:])
}

// invalidatePattern invalidates all cache entries matching a pattern
func invalidatePattern(ctx context.Context, client *redis.Client, pattern string) error {
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return client.Del(ctx, keys...).Err()
	}
	return nil
}

// ConditionalCache returns a middleware that supports conditional caching
func ConditionalCache(client *redis.Client, prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check If-None-Match header
		etag := c.GetHeader("If-None-Match")
		if etag != "" {
			key := prefix + "etag:" + etag
			if exists, _ := client.Exists(c.Request.Context(), key).Result(); exists > 0 {
				c.AbortWithStatus(http.StatusNotModified)
				return
			}
		}

		c.Next()
	}
}

