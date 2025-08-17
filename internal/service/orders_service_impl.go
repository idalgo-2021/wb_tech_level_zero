package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"wb_tech_level_zero/internal/config"
	"wb_tech_level_zero/internal/delivery/kafkadelivery"
	"wb_tech_level_zero/internal/orders"
	"wb_tech_level_zero/pkg/logger"

	"go.uber.org/zap"
)

const orderCachePrefix = "order:"

type ordersService struct {
	cfg   *config.Config
	repo  OrdersRepository
	cache OrdersCache
	wg    *sync.WaitGroup
	log   logger.Logger
}

func NewOrdersService(cfg *config.Config, repo OrdersRepository, cache OrdersCache, wg *sync.WaitGroup, log logger.Logger) OrdersService {
	return &ordersService{
		cfg:   cfg,
		repo:  repo,
		cache: cache,
		wg:    wg,
		log:   log,
	}
}

func (s *ordersService) GetOrderByUID(ctx context.Context, orderUID string) (*orders.Order, error) {
	key := orderCachePrefix + orderUID
	cached, err := s.cache.Get(ctx, key)
	if err != nil {
		s.log.Warn(ctx, "Failed to get order from cache", zap.String("key", key), zap.Error(err))
	}
	if cached != nil {
		return cached, nil
	}

	dbOrder, err := s.repo.GetOrderByUID(ctx, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order from repository: %w", err)
	}

	s.asyncCacheOrder(dbOrder)
	return dbOrder, nil

}

func (s *ordersService) GetOrders(ctx context.Context, params GetOrdersParams) ([]*orders.Order, int, error) {

	// TO DO: Переделать после уточнения требований(вызывать кэш, или удалить вообще)

	offset := (params.Page - 1) * params.Limit
	return s.repo.GetOrders(ctx, params.Limit, offset)
}

func (s *ordersService) ProcessEventOrder(ctx context.Context, eo *kafkadelivery.EventOrder) error {
	cached, _ := s.cache.Get(ctx, orderCachePrefix+eo.OrderUID)
	if cached != nil {
		s.log.Info(ctx, "Order already exists (found in cache), skipping", zap.String("order_uid", eo.OrderUID))
		return orders.ErrOrderAlreadyExists
	}

	order := mapEventOrderToDomain(eo)

	err := s.repo.SaveOrder(ctx, &order)
	if err != nil {
		if errors.Is(err, orders.ErrOrderAlreadyExists) {
			s.log.Info(ctx, "Order already exists (found in DB), skipping save", zap.String("order_uid", order.OrderUID))

			s.asyncCacheOrder(&order)
			return orders.ErrOrderAlreadyExists
		}
		return fmt.Errorf("failed to save order: %w", err)
	}

	s.asyncCacheOrder(&order)
	return nil
}

func (s *ordersService) WarmOrdersCache(ctx context.Context) error {
	s.log.Info(ctx, "Warming up cache...")

	// TO DO: Логика прогрева кэша(переработать после уточнения требований)

	orders, _, err := s.repo.GetOrders(ctx, s.cfg.CacheWarmupSize, 0)
	if err != nil {
		return fmt.Errorf("failed to load orders for warmup: %w", err)
	}

	failed := 0
	for _, order := range orders {
		if err := s.cacheOrder(ctx, order); err != nil {
			failed++
		}
	}
	s.log.Info(ctx, "Cache warmup completed", zap.Int("count", len(orders)), zap.Int("failed", failed))

	return nil
}

// //////////////

func (s *ordersService) asyncCacheOrder(order *orders.Order) {
	key := orderCachePrefix + order.OrderUID
	s.asyncCacheWithData(key, order)
}

func (s *ordersService) asyncCacheWithData(key string, order *orders.Order) {

	// TO DO: М,б.переделать логирование, канал ошибок и т.п.(для асинхронных операций)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ctxCache, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := s.cache.Set(ctxCache, key, order); err != nil {
			s.log.Warn(ctxCache, "Async cache order failed", zap.String("key", key), zap.Error(err))
		}

	}()
}

func (s *ordersService) cacheOrder(ctx context.Context, order *orders.Order) error {
	key := orderCachePrefix + order.OrderUID
	ctxCache, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.cache.Set(ctxCache, key, order); err != nil {
		s.log.Warn(ctx, "Failed to set order in cache", zap.String("order_uid", order.OrderUID), zap.Error(err))
		return err
	}

	return nil
}
