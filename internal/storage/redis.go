package storage

import (
	"context"
	"time"

	"rate-limiting-service/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage initializes and returns a RedisStorage instance.
func InitRedisStorage() *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     config.REDIS_ADDRESS,
		Username: config.REDIS_USERNAME,
		Password: config.REDIS_PASSWORD,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic("redis not connected: " + err.Error())
	}
	return &RedisStorage{
		client: client,
	}
}

// Set sets a key with a given value and TTL (time to live).
func (r *RedisStorage) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Get retrieves the value of a given key.
func (r *RedisStorage) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Incr increments the integer value of a key by one.
func (r *RedisStorage) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// TTL returns the time to live for a key.
func (r *RedisStorage) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Del deletes one or more keys.
func (r *RedisStorage) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if the key exists.
func (r *RedisStorage) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}
