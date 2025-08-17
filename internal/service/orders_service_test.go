package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"wb_tech_level_zero/internal/config"
	"wb_tech_level_zero/internal/delivery/kafkadelivery"
	"wb_tech_level_zero/internal/orders"

	"go.uber.org/zap"
)

type mockRepo struct {
	saveCalled bool
	saveErr    error
	getOrder   *orders.Order
	getOrders  []*orders.Order
	getErr     error
}

func (m *mockRepo) SaveOrder(ctx context.Context, o *orders.Order) error {
	m.saveCalled = true
	return m.saveErr
}

func (m *mockRepo) GetOrderByUID(ctx context.Context, uid string) (*orders.Order, error) {
	return m.getOrder, m.getErr
}

func (m *mockRepo) GetOrders(ctx context.Context, limit, offset int) ([]*orders.Order, int, error) {
	return m.getOrders, len(m.getOrders), m.getErr
}

/////////////////////////////

type mockCache struct {
	data   map[string]*orders.Order
	setErr error
	getErr error
}

func (m *mockCache) Get(ctx context.Context, key string) (*orders.Order, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.data[key], nil
}

func (m *mockCache) Set(ctx context.Context, key string, value *orders.Order) error {
	if m.setErr != nil {
		return m.setErr
	}
	if m.data == nil {
		m.data = map[string]*orders.Order{}
	}
	m.data[key] = value
	return nil
}

type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, msg string, fields ...zap.Field)  {}
func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...zap.Field)  {}
func (m *mockLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {}
func (m *mockLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {}

//////////////////////////////

func TestGetOrderByUID(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	order := &orders.Order{OrderUID: "o1"}
	logger := &mockLogger{}
	cfg := &config.Config{}

	t.Run("success: get from cache", func(t *testing.T) {
		cache := &mockCache{data: map[string]*orders.Order{"order:o1": order}}
		repo := &mockRepo{}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		got, err := svc.GetOrderByUID(ctx, "o1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if got.OrderUID != "o1" {
			t.Errorf("expected orderUID o1, got %s", got.OrderUID)
		}
	})

	t.Run("success: get from repo when cache is empty", func(t *testing.T) {
		cache := &mockCache{data: map[string]*orders.Order{}}
		repo := &mockRepo{getOrder: order}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		got, err := svc.GetOrderByUID(ctx, "o1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if got.OrderUID != "o1" {
			t.Errorf("expected orderUID o1, got %s", got.OrderUID)
		}
		// Проверяем, что после получения из репо, заказ попал в кэш
		wg.Wait()
		cachedVal, _ := cache.Get(ctx, "order:o1")
		if cachedVal == nil || cachedVal.OrderUID != "o1" {
			t.Error("order was not cached after fetching from repo")
		}
	})

	t.Run("error: not found in repo", func(t *testing.T) {
		cache := &mockCache{data: map[string]*orders.Order{}}
		repo := &mockRepo{getErr: orders.ErrOrderNotFound}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		_, err := svc.GetOrderByUID(ctx, "o1")
		if !errors.Is(err, orders.ErrOrderNotFound) {
			t.Errorf("expected ErrOrderNotFound, got %v", err)
		}
	})
}

func TestProcessEventOrder(t *testing.T) {
	eventOrder := &kafkadelivery.EventOrder{
		OrderUID: "new-order-123",
	}
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	logger := &mockLogger{}
	cfg := &config.Config{}

	t.Run("success: new order is saved and cached", func(t *testing.T) {
		repo := &mockRepo{}
		cache := &mockCache{}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		err := svc.ProcessEventOrder(ctx, eventOrder)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !repo.saveCalled {
			t.Error("repo.SaveOrder was not called")
		}

		wg.Wait()
		cachedVal, _ := cache.Get(ctx, "order:"+eventOrder.OrderUID)
		if cachedVal == nil {
			t.Error("order was not cached after saving")
		}
	})

	t.Run("duplicate: order already in cache", func(t *testing.T) {
		repo := &mockRepo{}
		order := &orders.Order{OrderUID: eventOrder.OrderUID}
		cache := &mockCache{data: map[string]*orders.Order{"order:" + eventOrder.OrderUID: order}}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		err := svc.ProcessEventOrder(ctx, eventOrder)
		if !errors.Is(err, orders.ErrOrderAlreadyExists) {
			t.Errorf("expected ErrOrderAlreadyExists, got %v", err)
		}

		if repo.saveCalled {
			t.Error("repo.SaveOrder should not be called if order is in cache")
		}
	})

	t.Run("duplicate: order already in db, cache is updated", func(t *testing.T) {
		repo := &mockRepo{saveErr: orders.ErrOrderAlreadyExists}
		cache := &mockCache{}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		err := svc.ProcessEventOrder(ctx, eventOrder)
		if !errors.Is(err, orders.ErrOrderAlreadyExists) {
			t.Errorf("expected ErrOrderAlreadyExists, got %v", err)
		}

		wg.Wait()
		cachedVal, _ := cache.Get(ctx, "order:"+eventOrder.OrderUID)
		if cachedVal == nil {
			t.Error("order should be cached even if it already exists in db (cache self-healing)")
		}
	})

	t.Run("error: failed to save order in db", func(t *testing.T) {
		dbErr := errors.New("db is down")
		repo := &mockRepo{saveErr: dbErr}
		cache := &mockCache{}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		err := svc.ProcessEventOrder(ctx, eventOrder)
		if !errors.Is(err, dbErr) {
			t.Errorf("expected wrapped db error, got %v", err)
		}

		if len(cache.data) > 0 {
			t.Error("cache should be empty on db save failure")
		}
	})
}

func TestWarmOrdersCache(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	logger := &mockLogger{}
	cfg := &config.Config{CacheWarmupSize: 2}
	ordersToWarm := []*orders.Order{
		{OrderUID: "warm1"},
		{OrderUID: "warm2"},
	}

	t.Run("success: all orders are cached", func(t *testing.T) {
		repo := &mockRepo{getOrders: ordersToWarm}
		cache := &mockCache{}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		err := svc.WarmOrdersCache(ctx)
		if err != nil {
			t.Fatalf("unexpected error during cache warmup: %v", err)
		}

		if len(cache.data) != 2 {
			t.Errorf("expected 2 items in cache, got %d", len(cache.data))
		}
		if _, ok := cache.data["order:warm1"]; !ok {
			t.Error("order warm1 not found in cache")
		}
	})

	t.Run("error: repo fails to get orders", func(t *testing.T) {
		repoErr := errors.New("db is down")
		repo := &mockRepo{getErr: repoErr}
		cache := &mockCache{}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		err := svc.WarmOrdersCache(ctx)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected wrapped db error, got %v", err)
		}
		if len(cache.data) > 0 {
			t.Error("cache should be empty when repo fails")
		}
	})

	t.Run("partial success: cache set fails", func(t *testing.T) {
		repo := &mockRepo{getOrders: ordersToWarm}
		// This mock will fail on every Set call
		cache := &mockCache{setErr: errors.New("redis is flaky")}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		// The function should not return an error, as it only logs cache failures during warmup
		err := svc.WarmOrdersCache(ctx)
		if err != nil {
			t.Fatalf("expected no error from WarmOrdersCache on cache.Set failure, got %v", err)
		}
	})
}

func TestGetOrders(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	logger := &mockLogger{}
	cfg := &config.Config{}
	ordersList := []*orders.Order{
		{OrderUID: "o1"},
		{OrderUID: "o2"},
	}

	t.Run("success: get orders list", func(t *testing.T) {
		repo := &mockRepo{getOrders: ordersList}
		cache := &mockCache{}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		params := GetOrdersParams{Page: 1, Limit: 10}
		got, total, err := svc.GetOrders(ctx, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(got) != 2 {
			t.Errorf("expected 2 orders, got %d", len(got))
		}
	})

	t.Run("error: repo fails", func(t *testing.T) {
		repoErr := errors.New("db is down")
		repo := &mockRepo{getErr: repoErr}
		cache := &mockCache{}
		svc := NewOrdersService(cfg, repo, cache, wg, logger)

		params := GetOrdersParams{Page: 1, Limit: 10}
		_, _, err := svc.GetOrders(ctx, params)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected db error, got %v", err)
		}
	})
}
