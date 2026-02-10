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
	productCachePrefix = "catalog:product:"
	productCacheTTL    = 30 * time.Minute
)

// ProductCache handles product caching
type ProductCache struct {
	client *redis.Client
}

// NewProductCache creates a new ProductCache
func NewProductCache(client *redis.Client) *ProductCache {
	return &ProductCache{client: client}
}

// Get retrieves a product from cache
func (c *ProductCache) Get(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	key := c.key(id)
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var product models.Product
	if err := json.Unmarshal([]byte(data), &product); err != nil {
		return nil, err
	}

	return &product, nil
}

// Set stores a product in cache
func (c *ProductCache) Set(ctx context.Context, product *models.Product) error {
	key := c.key(product.ID)
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, productCacheTTL).Err()
}

// Delete removes a product from cache
func (c *ProductCache) Delete(ctx context.Context, id uuid.UUID) error {
	key := c.key(id)
	return c.client.Del(ctx, key).Err()
}

// GetBySKU retrieves a product by SKU from cache
func (c *ProductCache) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	key := productCachePrefix + "sku:" + sku
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var product models.Product
	if err := json.Unmarshal([]byte(data), &product); err != nil {
		return nil, err
	}

	return &product, nil
}

// SetBySKU stores a product with SKU key
func (c *ProductCache) SetBySKU(ctx context.Context, product *models.Product) error {
	key := productCachePrefix + "sku:" + product.SKU
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, productCacheTTL).Err()
}

// GetMulti retrieves multiple products from cache
func (c *ProductCache) GetMulti(ctx context.Context, ids []uuid.UUID) ([]*models.Product, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = c.key(id)
	}

	results, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	products := make([]*models.Product, 0, len(results))
	for _, result := range results {
		if result == nil {
			continue
		}
		data, ok := result.(string)
		if !ok {
			continue
		}
		var product models.Product
		if err := json.Unmarshal([]byte(data), &product); err == nil {
			products = append(products, &product)
		}
	}

	return products, nil
}

// InvalidateAll invalidates all product cache
func (c *ProductCache) InvalidateAll(ctx context.Context) error {
	keys, err := c.client.Keys(ctx, productCachePrefix+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (c *ProductCache) key(id uuid.UUID) string {
	return fmt.Sprintf("%s%s", productCachePrefix, id.String())
}

