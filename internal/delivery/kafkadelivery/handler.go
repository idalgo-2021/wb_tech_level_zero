////////////////////////////
// internal/delivery/kafkadelivery/handler.go

package kafkadelivery

import (
	"context"

	"wb_tech_level_zero/pkg/logger"

	kafkaGo "github.com/segmentio/kafka-go"
)

type OrdersService interface {
	// GetOrderByUID(ctx context.Context, orderUID string) (*orders.Order, error)
}

type Handler struct {
	orderService OrdersService
	logger       logger.Logger
}

func NewHandler(orderService OrdersService, logger logger.Logger) *Handler {
	return &Handler{
		orderService: orderService,
		logger:       logger,
	}
}

func (h Handler) HandleMessage(ctx context.Context, msg kafkaGo.Message) error {
	h.logger.Info(ctx, "Received Kafka message: "+string(msg.Value))

	// var order orders.Order

	// eventOrder, err := ParseEventOrder(msg.Value)
	// if err != nil {
	// 	h.logger.Error(ctx, "Failed to parse message", zap.Error(err))
	// 	return err
	// }

	// // Валидируем
	// if err := ValidateEventOrder(eventOrder); err != nil {
	// 	h.logger.Error(ctx, "Validation failed", zap.Error(err))
	// 	return err
	// }

	// if err := h.service.ProcessOrder(ctx, &order); err != nil {
	// 	h.logger.Error(ctx, "Failed to process order in service layer", err)
	// 	return err
	// }

	// h.logger.Info(ctx, "Order processed successfully", "order_uid", order.OrderUID)
	return nil
}
