package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ecomm-stack/catalog-service/internal/infrastructure/db"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// Health performs health check
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  make(map[string]string),
	}

	// Check database
	if err := db.HealthCheck(ctx, h.db); err != nil {
		response.Services["database"] = "unhealthy: " + err.Error()
		response.Status = "unhealthy"
	} else {
		response.Services["database"] = "healthy"
	}

	// Check Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		response.Services["redis"] = "unhealthy: " + err.Error()
		response.Status = "unhealthy"
	} else {
		response.Services["redis"] = "healthy"
	}

	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Liveness handles liveness probe
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Readiness handles readiness probe
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	// Check database connection
	if err := db.HealthCheck(ctx, h.db); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  err.Error(),
		})
		return
	}

	// Check Redis connection
	if err := h.redis.Ping(ctx).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

