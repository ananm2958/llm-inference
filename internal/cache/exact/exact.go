package exact

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourorg/llm-gateway/internal/providers"
)

const defaultTTL = 24 * time.Hour

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(addr, password string, db int) *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Cache{client: rdb, ttl: defaultTTL}
}

func (c *Cache) Get(ctx context.Context, key string) (*providers.ChatResponse, error) {
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var resp providers.ChatResponse
	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Cache) Set(ctx context.Context, key string, resp *providers.ChatResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
