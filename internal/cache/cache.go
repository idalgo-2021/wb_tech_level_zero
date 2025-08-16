package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wb_tech_level_zero/internal/orders"

	"github.com/redis/go-redis/v9"
)

type OrdersCache struct {
	cacheClient *redis.Client
	ttl         time.Duration
}

type CacheConfig struct {
	TTL int
}

func NewOrdersCache(client *redis.Client, cfg CacheConfig) *OrdersCache {
	return &OrdersCache{
		cacheClient: client,
		ttl:         time.Duration(cfg.TTL) * time.Minute,
	}
}

func (r *OrdersCache) Get(ctx context.Context, key string) (*orders.Order, error) {
	val, err := r.cacheClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var order orders.Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached order: %w", err)
	}

	return &order, nil
}

func (r *OrdersCache) Set(ctx context.Context, key string, order *orders.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order for cache: %w", err)
	}

	if err := r.cacheClient.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}
