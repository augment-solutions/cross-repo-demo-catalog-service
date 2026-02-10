package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	APIKeys        []string
	JWTSecret      string
	SkipPaths      []string
	EnableAPIKey   bool
	EnableJWT      bool
}

// Auth returns an authentication middleware
func Auth(cfg AuthConfig, logger *zap.Logger) gin.HandlerFunc {
	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	apiKeys := make(map[string]bool)
	for _, key := range cfg.APIKeys {
		apiKeys[key] = true
	}

	return func(c *gin.Context) {
		// Skip authentication for certain paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Check API key if enabled
		if cfg.EnableAPIKey {
			apiKey := c.GetHeader("X-API-Key")
			if apiKey != "" && apiKeys[apiKey] {
				c.Set("auth_method", "api_key")
				c.Next()
				return
			}
		}

		// Check JWT if enabled
		if cfg.EnableJWT {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					token := parts[1]
					if validateJWT(token, cfg.JWTSecret) {
						c.Set("auth_method", "jwt")
						c.Next()
						return
					}
				}
			}
		}

		// No valid authentication found
		logger.Warn("unauthorized request",
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
	}
}

// validateJWT validates a JWT token (placeholder implementation)
func validateJWT(token, secret string) bool {
	// In production, implement proper JWT validation
	return token != "" && secret != ""
}

// CORS returns a CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-API-Key, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID, X-Trace-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimit returns a rate limiting middleware
func RateLimit(requestsPerSecond int) gin.HandlerFunc {
	// Simple rate limiter using token bucket
	return func(c *gin.Context) {
		// In production, implement proper rate limiting with Redis
		c.Next()
	}
}

// RequireRole returns a middleware that requires a specific role
func RequireRole(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool)
	for _, role := range roles {
		roleSet[role] = true
	}

	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "role required",
			})
			return
		}

		if !roleSet[userRole.(string)] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

