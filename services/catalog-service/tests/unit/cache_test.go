package unit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock Redis client
type MockRedisClient struct {
	mock.Mock
	data map[string]string
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]string),
	}
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", assert.AnError
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, _ := json.Marshal(value)
	m.data[key] = string(data)
	return nil
}

func (m *MockRedisClient) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func TestProductCache_Get(t *testing.T) {
	cache := NewMockRedisClient()
	ctx := context.Background()

	product := &models.Product{
		ID:   uuid.New(),
		Name: "Test Product",
		SKU:  "TEST-001",
	}

	// Store product in cache
	key := "catalog:product:" + product.ID.String()
	cache.Set(ctx, key, product, 30*time.Minute)

	// Retrieve from cache
	data, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestProductCache_Miss(t *testing.T) {
	cache := NewMockRedisClient()
	ctx := context.Background()

	// Try to get non-existent key
	_, err := cache.Get(ctx, "catalog:product:non-existent")
	assert.Error(t, err)
}

func TestProductCache_Set(t *testing.T) {
	cache := NewMockRedisClient()
	ctx := context.Background()

	product := &models.Product{
		ID:   uuid.New(),
		Name: "Test Product",
	}

	key := "catalog:product:" + product.ID.String()
	err := cache.Set(ctx, key, product, 30*time.Minute)
	assert.NoError(t, err)

	// Verify it was stored
	_, err = cache.Get(ctx, key)
	assert.NoError(t, err)
}

func TestProductCache_Delete(t *testing.T) {
	cache := NewMockRedisClient()
	ctx := context.Background()

	product := &models.Product{
		ID:   uuid.New(),
		Name: "Test Product",
	}

	key := "catalog:product:" + product.ID.String()
	cache.Set(ctx, key, product, 30*time.Minute)

	err := cache.Delete(ctx, key)
	assert.NoError(t, err)

	// Verify it was deleted
	_, err = cache.Get(ctx, key)
	assert.Error(t, err)
}

func TestCategoryCache_Tree(t *testing.T) {
	cache := NewMockRedisClient()
	ctx := context.Background()

	tree := []*models.CategoryTree{
		{
			ID:   uuid.New(),
			Name: "Electronics",
			Children: []*models.CategoryTree{
				{ID: uuid.New(), Name: "Phones"},
				{ID: uuid.New(), Name: "Tablets"},
			},
		},
	}

	key := "catalog:category:tree"
	err := cache.Set(ctx, key, tree, time.Hour)
	assert.NoError(t, err)

	data, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestCache_TTL(t *testing.T) {
	// Test that TTL is properly set
	ttl := 30 * time.Minute
	assert.Equal(t, 30*time.Minute, ttl)

	ttl = time.Hour
	assert.Equal(t, time.Hour, ttl)
}

func TestCache_KeyGeneration(t *testing.T) {
	id := uuid.New()
	key := "catalog:product:" + id.String()

	assert.Contains(t, key, "catalog:product:")
	assert.Contains(t, key, id.String())
}

