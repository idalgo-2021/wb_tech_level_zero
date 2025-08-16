package service

import (
	"context"
	"wb_tech_level_zero/internal/orders"
)

type OrdersRepository interface {
	GetOrderByUID(ctx context.Context, orderUID string) (*orders.Order, error)
	SaveOrder(ctx context.Context, order *orders.Order) error

	GetOrders(ctx context.Context, limit, offset int) ([]*orders.Order, int, error)
}
