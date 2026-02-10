package cache

import (
	"context"
	"time"

	"github.com/ecomm-stack/catalog-service/internal/config"
	"github.com/go-redis/redis/v8"
)

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.Config) *redis.Client {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		opt = &redis.Options{
			Addr: cfg.RedisURL,
		}
	}

	client := redis.NewClient(opt)
	return client
}

// Ping checks Redis connection
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}

// Set sets a value with TTL
func Set(ctx context.Context, client *redis.Client, key string, value interface{}, ttl time.Duration) error {
	return client.Set(ctx, key, value, ttl).Err()
}

// Get gets a value
func Get(ctx context.Context, client *redis.Client, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

// Delete deletes a key
func Delete(ctx context.Context, client *redis.Client, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func Exists(ctx context.Context, client *redis.Client, key string) (bool, error) {
	result, err := client.Exists(ctx, key).Result()
	return result > 0, err
}

// SetNX sets a value if not exists
func SetNX(ctx context.Context, client *redis.Client, key string, value interface{}, ttl time.Duration) (bool, error) {
	return client.SetNX(ctx, key, value, ttl).Result()
}

// Incr increments a counter
func Incr(ctx context.Context, client *redis.Client, key string) (int64, error) {
	return client.Incr(ctx, key).Result()
}

// HSet sets a hash field
func HSet(ctx context.Context, client *redis.Client, key string, values ...interface{}) error {
	return client.HSet(ctx, key, values...).Err()
}

// HGet gets a hash field
func HGet(ctx context.Context, client *redis.Client, key, field string) (string, error) {
	return client.HGet(ctx, key, field).Result()
}

// HGetAll gets all hash fields
func HGetAll(ctx context.Context, client *redis.Client, key string) (map[string]string, error) {
	return client.HGetAll(ctx, key).Result()
}

// HDel deletes hash fields
func HDel(ctx context.Context, client *redis.Client, key string, fields ...string) error {
	return client.HDel(ctx, key, fields...).Err()
}

// Keys gets keys matching pattern
func Keys(ctx context.Context, client *redis.Client, pattern string) ([]string, error) {
	return client.Keys(ctx, pattern).Result()
}

// TTL gets the TTL of a key
func TTL(ctx context.Context, client *redis.Client, key string) (time.Duration, error) {
	return client.TTL(ctx, key).Result()
}

// Expire sets TTL on a key
func Expire(ctx context.Context, client *redis.Client, key string, ttl time.Duration) error {
	return client.Expire(ctx, key, ttl).Err()
}

