////////////////////////////
// internal/service/service.go

package service

import (
	"context"
	"wb_tech_level_zero/internal/config"

	"wb_tech_level_zero/internal/orders"
)

type OrdersRepository interface {
	GetOrderByUID(ctx context.Context, orderUID string) (*orders.Order, error)
	GetOrders(ctx context.Context, limit, offset int) ([]*orders.Order, int, error)
}

type ordersService struct {
	repo OrdersRepository
	cfg  *config.Config
}

func NewOrdersService(cfg *config.Config, repo OrdersRepository) *ordersService {
	return &ordersService{
		cfg:  cfg,
		repo: repo,
	}
}

func (s *ordersService) GetOrderByUID(ctx context.Context, orderUID string) (*orders.Order, error) {

	// TODO: добавить логику кэширования через Redis
	// 1. Проверить кэш по ключу "order:id"
	// 2. Если есть в кэше — вернуть
	// 3. Если нет — взять из repo, положить в кэш, вернуть
	// return s.repo.GetOrderByID(ctx, id)

	// dbOrder, err := s.repo.GetOrderByUID(ctx, orderUID)
	// if err != nil {
	// 	return nil, err
	// }
	// return dbOrder, nil

	return s.repo.GetOrderByUID(ctx, orderUID)
}

type GetOrdersParams struct {
	Page  int
	Limit int
}

func (s *ordersService) GetOrders(ctx context.Context, params GetOrdersParams) ([]*orders.Order, int, error) {

	// 	// TODO: добавить логику кэширования через Redis
	// 	// 1. Проверить кэш по ключу "orders:limit:offset"
	// 	// 2. Если есть в кэше — вернуть
	// 	// 3. Если нет — взять из repo, положить в кэш, вернуть
	// 	// return s.repo.GetOrders(ctx, limit, offset)

	offset := (params.Page - 1) * params.Limit

	return s.repo.GetOrders(ctx, params.Limit, offset)

}
