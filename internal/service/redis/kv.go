package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// ErrNotFound indicates a missing key in Redis.
var ErrNotFound = errors.New("redis: key not found")

// KV defines a thin key-value surface for Redis operations.
// Pooling/retry/health/logging are handled by infra layer.
type KV interface {
	Ping(ctx context.Context) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) (int64, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
}

// clientKV is a thin adapter around go-redis Client implementing KV.
type clientKV struct {
	client *goredis.Client
}

// NewKVFromClient adapts a go-redis client to the KV interface.
func NewKVFromClient(client *goredis.Client) KV {
	if client == nil {
		return nil
	}
	return &clientKV{client: client}
}

func (c *clientKV) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *clientKV) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", ErrNotFound
		}
		return "", err
	}
	return val, nil
}

func (c *clientKV) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *clientKV) Del(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Del(ctx, keys...).Result()
}

func (c *clientKV) Keys(ctx context.Context, pattern string) ([]string, error) {
	const scanCount = 1000
	var cursor uint64
	var out []string
	// Use SCAN to avoid blocking Redis for large keyspaces
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, scanCount).Result()
		if err != nil {
			return nil, err
		}
		out = append(out, keys...)
		cursor = next
		if cursor == 0 {
			break
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	return out, nil
}
