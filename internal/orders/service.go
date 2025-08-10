// service.go
package orders

import (
	"context"
	"wb_tech_level_zero/internal/config"
	"wb_tech_level_zero/internal/models"
)

type OrdersRepository interface {
	GetOrderByUID(ctx context.Context, orderUID string) (*models.Order, error)
	GetOrders(ctx context.Context, limit, offset int) ([]*models.Order, int, error)
	// GetOrderByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
}

type OrdersService struct {
	repo OrdersRepository
	cfg  *config.Config
}

func NewOrdersService(repo OrdersRepository, cfg *config.Config) *OrdersService {
	return &OrdersService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *OrdersService) GetOrders(ctx context.Context, dto GetOrdersRequest) (*OrdersResponse, error) {

	// TODO: добавить логику кэширования через Redis
	// 1. Проверить кэш по ключу "orders:limit:offset"
	// 2. Если есть в кэше — вернуть
	// 3. Если нет — взять из repo, положить в кэш, вернуть
	// return s.repo.GetOrders(ctx, limit, offset)

	offset := (dto.Page - 1) * dto.Limit

	dbOrders, total, err := s.repo.GetOrders(ctx, dto.Limit, offset)
	if err != nil {
		return nil, err
	}

	responseDTOs := make([]*OrderDTO, 0, len(dbOrders))
	for _, o := range dbOrders {
		responseDTOs = append(responseDTOs, &OrderDTO{
			// ID:                o.ID,
			OrderUID:          o.OrderUID,
			TrackNumber:       o.TrackNumber,
			Entry:             o.Entry,
			Delivery:          DeliveryDTO(o.Delivery),
			Payment:           PaymentDTO(o.Payment),
			Items:             toItemDTOs(o.Items),
			Locale:            o.Locale,
			InternalSignature: o.InternalSignature,
			CustomerID:        o.CustomerID,
			DeliveryService:   o.DeliveryService,
			Shardkey:          o.Shardkey,
			SmID:              o.SmID,
			DateCreated:       o.DateCreated,
			OofShard:          o.OofShard,
		})

	}

	response := &OrdersResponse{
		Orders: responseDTOs,
		Total:  total,
		Page:   dto.Page,
	}
	return response, nil

}

func toItemDTOs(items []models.Item) []ItemDTO {
	result := make([]ItemDTO, 0, len(items))
	for _, i := range items {
		result = append(result, ItemDTO{
			ChrtID:      i.ChrtID,
			TrackNumber: i.TrackNumber,
			Price:       i.Price,
			Rid:         i.Rid,
			Name:        i.Name,
			Sale:        i.Sale,
			Size:        i.Size,
			TotalPrice:  i.TotalPrice,
			NmID:        i.NmID,
			Brand:       i.Brand,
			Status:      i.Status,
		})
	}
	return result
}

func (s *OrdersService) GetOrderByUID(ctx context.Context, orderUID string) (*OrderDTO, error) {

	// TODO: добавить логику кэширования через Redis
	// 1. Проверить кэш по ключу "order:id"
	// 2. Если есть в кэше — вернуть
	// 3. Если нет — взять из repo, положить в кэш, вернуть
	// return s.repo.GetOrderByID(ctx, id)

	dbOrder, err := s.repo.GetOrderByUID(ctx, orderUID)
	if err != nil {
		return nil, err
	}

	return &OrderDTO{
		OrderUID:          dbOrder.OrderUID,
		TrackNumber:       dbOrder.TrackNumber,
		Entry:             dbOrder.Entry,
		Delivery:          DeliveryDTO(dbOrder.Delivery),
		Payment:           PaymentDTO(dbOrder.Payment),
		Items:             toItemDTOs(dbOrder.Items),
		Locale:            dbOrder.Locale,
		InternalSignature: dbOrder.InternalSignature,
		CustomerID:        dbOrder.CustomerID,
		DeliveryService:   dbOrder.DeliveryService,
		Shardkey:          dbOrder.Shardkey,
		SmID:              dbOrder.SmID,
		DateCreated:       dbOrder.DateCreated,
		OofShard:          dbOrder.OofShard,
	}, nil
}

// func (s *OrdersService) GetOrderByID(ctx context.Context, id uuid.UUID) (*OrderDTO, error) {
//
// 	// TODO: добавить логику кэширования через Redis
// 	// 1. Проверить кэш по ключу "order:id"
// 	// 2. Если есть в кэше — вернуть
// 	// 3. Если нет — взять из repo, положить в кэш, вернуть
// 	// return s.repo.GetOrderByID(ctx, id)
//
// 	dbOrder, err := s.repo.GetOrderByID(ctx, id)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &OrderDTO{
// 		// ID:             dbOrder.ID,
// 		OrderUID:          dbOrder.OrderUID,
// 		TrackNumber:       dbOrder.TrackNumber,
// 		Entry:             dbOrder.Entry,
// 		Delivery:          DeliveryDTO(dbOrder.Delivery),
// 		Payment:           PaymentDTO(dbOrder.Payment),
// 		Items:             toItemDTOs(dbOrder.Items),
// 		Locale:            dbOrder.Locale,
// 		InternalSignature: dbOrder.InternalSignature,
// 		CustomerID:        dbOrder.CustomerID,
// 		DeliveryService:   dbOrder.DeliveryService,
// 		Shardkey:          dbOrder.Shardkey,
// 		SmID:              dbOrder.SmID,
// 		DateCreated:       dbOrder.DateCreated,
// 		OofShard:          dbOrder.OofShard,
// 	}, nil
//
// }
