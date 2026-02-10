package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	categoryCachePrefix = "catalog:category:"
	categoryTreeKey     = "catalog:category:tree"
	categoryCacheTTL    = 1 * time.Hour
)

// CategoryCache handles category caching
type CategoryCache struct {
	client *redis.Client
}

// NewCategoryCache creates a new CategoryCache
func NewCategoryCache(client *redis.Client) *CategoryCache {
	return &CategoryCache{client: client}
}

// Get retrieves a category from cache
func (c *CategoryCache) Get(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	key := c.key(id)
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var category models.Category
	if err := json.Unmarshal([]byte(data), &category); err != nil {
		return nil, err
	}

	return &category, nil
}

// Set stores a category in cache
func (c *CategoryCache) Set(ctx context.Context, category *models.Category) error {
	key := c.key(category.ID)
	data, err := json.Marshal(category)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, categoryCacheTTL).Err()
}

// Delete removes a category from cache
func (c *CategoryCache) Delete(ctx context.Context, id uuid.UUID) error {
	key := c.key(id)
	return c.client.Del(ctx, key).Err()
}

// GetTree retrieves the category tree from cache
func (c *CategoryCache) GetTree(ctx context.Context) ([]*models.CategoryTree, error) {
	data, err := c.client.Get(ctx, categoryTreeKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var tree []*models.CategoryTree
	if err := json.Unmarshal([]byte(data), &tree); err != nil {
		return nil, err
	}

	return tree, nil
}

// SetTree stores the category tree in cache
func (c *CategoryCache) SetTree(ctx context.Context, tree []*models.CategoryTree) error {
	data, err := json.Marshal(tree)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, categoryTreeKey, data, categoryCacheTTL).Err()
}

// InvalidateTree invalidates the category tree cache
func (c *CategoryCache) InvalidateTree(ctx context.Context) error {
	return c.client.Del(ctx, categoryTreeKey).Err()
}

// GetBySlug retrieves a category by slug from cache
func (c *CategoryCache) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	key := categoryCachePrefix + "slug:" + slug
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var category models.Category
	if err := json.Unmarshal([]byte(data), &category); err != nil {
		return nil, err
	}

	return &category, nil
}

// SetBySlug stores a category with slug key
func (c *CategoryCache) SetBySlug(ctx context.Context, category *models.Category) error {
	key := categoryCachePrefix + "slug:" + category.Slug
	data, err := json.Marshal(category)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, categoryCacheTTL).Err()
}

// InvalidateAll invalidates all category cache
func (c *CategoryCache) InvalidateAll(ctx context.Context) error {
	keys, err := c.client.Keys(ctx, categoryCachePrefix+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (c *CategoryCache) key(id uuid.UUID) string {
	return fmt.Sprintf("%s%s", categoryCachePrefix, id.String())
}

