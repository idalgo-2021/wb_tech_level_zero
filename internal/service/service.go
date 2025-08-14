////////////////////////////
// internal/service/service.go

package service

import (
	"context"
	"errors"
	"wb_tech_level_zero/internal/config"
	"wb_tech_level_zero/internal/delivery/kafkadelivery"
	"wb_tech_level_zero/pkg/logger"

	"wb_tech_level_zero/internal/orders"

	"go.uber.org/zap"
)

type OrdersRepository interface {
	GetOrderByUID(ctx context.Context, orderUID string) (*orders.Order, error)
	GetOrders(ctx context.Context, limit, offset int) ([]*orders.Order, int, error)

	SaveOrder(ctx context.Context, order *orders.Order) error
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

	// TODO: добавить логику кэширования через Redis
	// 1. Проверить кэш по ключу "orders:limit:offset"
	// 2. Если есть в кэше — вернуть
	// 3. Если нет — взять из repo, положить в кэш, вернуть
	// return s.repo.GetOrders(ctx, limit, offset)

	offset := (params.Page - 1) * params.Limit

	return s.repo.GetOrders(ctx, params.Limit, offset)

}

////////

func (s *ordersService) ProcessEventOrder(ctx context.Context, eo *kafkadelivery.EventOrder) error {
	log := logger.GetLoggerFromCtx(ctx)

	order := s.mapEventOrderToDomain(eo)

	err := s.repo.SaveOrder(ctx, &order)
	if err != nil {
		if errors.Is(err, orders.ErrOrderAlreadyExists) {
			log.Info(ctx, "Order already exists, skipping", zap.String("order_uid", order.OrderUID))
			return nil
		}
		log.Error(ctx, "Failed to save order", zap.String("order_uid", order.OrderUID), zap.Error(err))
		return err
	}

	log.Info(ctx, "Order saved successfully", zap.String("order_uid", order.OrderUID))
	return nil
}

func (s *ordersService) mapEventOrderToDomain(eo *kafkadelivery.EventOrder) orders.Order {
	order := orders.Order{
		OrderUID:          eo.OrderUID,
		TrackNumber:       eo.TrackNumber,
		Entry:             eo.Entry,
		Locale:            eo.Locale,
		InternalSignature: eo.InternalSignature,
		CustomerID:        eo.CustomerID,
		DeliveryService:   eo.DeliveryService,
		Shardkey:          eo.Shardkey,
		SmID:              eo.SmID,
		DateCreated:       &eo.DateCreated,
		OofShard:          eo.OofShard,
		Delivery: orders.Delivery{
			Name:    eo.Delivery.Name,
			Phone:   eo.Delivery.Phone,
			Zip:     eo.Delivery.Zip,
			City:    eo.Delivery.City,
			Address: eo.Delivery.Address,
			Region:  eo.Delivery.Region,
			Email:   eo.Delivery.Email,
		},
		Payment: orders.Payment{
			Transaction:  eo.Payment.Transaction,
			RequestID:    eo.Payment.RequestID,
			Currency:     eo.Payment.Currency,
			Provider:     eo.Payment.Provider,
			Amount:       eo.Payment.Amount,
			PaymentDT:    eo.Payment.PaymentDT,
			Bank:         eo.Payment.Bank,
			DeliveryCost: eo.Payment.DeliveryCost,
			GoodsTotal:   eo.Payment.GoodsTotal,
			CustomFee:    eo.Payment.CustomFee,
		},
	}

	items := make([]orders.Item, len(eo.Items))
	for i, it := range eo.Items {
		items[i] = orders.Item{
			ChrtID:      it.ChrtID,
			TrackNumber: it.TrackNumber,
			Price:       it.Price,
			Rid:         it.Rid,
			Name:        it.Name,
			Sale:        it.Sale,
			Size:        it.Size,
			TotalPrice:  it.TotalPrice,
			NmID:        it.NmID,
			Brand:       it.Brand,
			Status:      it.Status,
		}
	}
	order.Items = items

	return order
}
