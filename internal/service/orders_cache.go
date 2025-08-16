package service

import (
	"context"
	"wb_tech_level_zero/internal/orders"
)

type OrdersCache interface {
	Get(ctx context.Context, key string) (*orders.Order, error)
	Set(ctx context.Context, key string, value *orders.Order) error
}
