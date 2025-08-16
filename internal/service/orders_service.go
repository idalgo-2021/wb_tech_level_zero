package service

import (
	"context"
	"wb_tech_level_zero/internal/delivery/kafkadelivery"
	"wb_tech_level_zero/internal/orders"
)

type OrdersService interface {
	WarmOrdersCache(ctx context.Context) error
	GetOrderByUID(ctx context.Context, uid string) (*orders.Order, error)
	ProcessEventOrder(ctx context.Context, eo *kafkadelivery.EventOrder) error

	GetOrders(ctx context.Context, params GetOrdersParams) ([]*orders.Order, int, error)
}
