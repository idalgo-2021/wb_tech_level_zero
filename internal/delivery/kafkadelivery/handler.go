////////////////////////////
// internal/delivery/kafkadelivery/handler.go

package kafkadelivery

import (
	"context"

	"wb_tech_level_zero/pkg/logger"

	"github.com/go-playground/validator/v10"
	kafkaGo "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type OrdersService interface {
	ProcessEventOrder(ctx context.Context, eventOrder *EventOrder) error
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

	eventOrder, err := ParseAndValidate(msg.Value)
	if err != nil {
		h.logger.Error(ctx, "Failed to parse or validate message", zap.Error(err))

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				h.logger.Error(ctx, "Validation error",
					zap.String("field", fe.Field()),
					zap.String("tag", fe.Tag()),
					zap.String("value", fe.Param()))
			}
		}
		return err
	}

	err = h.orderService.ProcessEventOrder(ctx, eventOrder)
	if err != nil {
		h.logger.Error(ctx, "Failed to process order in service layer", zap.Error(err))
		return err
	}

	h.logger.Info(ctx, "Order processed successfully, order_uid: ", zap.String("order_uid", eventOrder.OrderUID))

	return nil
}
