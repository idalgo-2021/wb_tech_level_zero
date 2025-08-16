package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"wb_tech_level_zero/internal/config"
	"wb_tech_level_zero/internal/dto"
	"wb_tech_level_zero/internal/orders"
	"wb_tech_level_zero/internal/service"
	"wb_tech_level_zero/pkg/logger"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type OrdersService interface {
	GetOrderByUID(ctx context.Context, orderUID string) (*orders.Order, error)
	GetOrders(ctx context.Context, params service.GetOrdersParams) ([]*orders.Order, int, error)
}

type Handlers struct {
	cfg          *config.Config
	orderService OrdersService
}

func NewHandlers(cfg *config.Config, orderService OrdersService) *Handlers {
	return &Handlers{
		cfg:          cfg,
		orderService: orderService,
	}
}

// @Summary Getting orders by UID
// @Description Getting orders by UID
// @Tags orders
// @Produce json
// @Param uid path string true "UID заказа"
// @Success 200 {object} dto.OrderDTO
// @Failure 404 {string} string "Not Found"
// @Router /order/{uid} [get]
func (h *Handlers) GetOrderByUID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromCtx(ctx)

	vars := mux.Vars(r)
	orderUID, ok := vars["order_uid"]
	if !ok || orderUID == "" {
		log.Error(ctx, "URL path parameter 'order_uid' is missing")
		h.writeErrorResponse(ctx, w, http.StatusBadRequest, "order_uid path parameter is required")
		return
	}

	dbOrder, err := h.orderService.GetOrderByUID(ctx, orderUID)
	if err != nil {
		if errors.Is(err, orders.ErrOrderNotFound) {
			log.Info(ctx, "Order not found by order_uid", zap.String("order_uid", orderUID))
			h.writeErrorResponse(ctx, w, http.StatusNotFound, "Order not found")
		} else {
			log.Error(ctx, "Failed to get order by order_uid",
				zap.Error(err),
				zap.String("order_uid", orderUID),
			)
			h.writeErrorResponse(ctx, w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	orderDTO := dto.OrderToDTO(dbOrder)
	h.writeJSONResponse(ctx, w, http.StatusOK, orderDTO)

}

func (h *Handlers) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromCtx(ctx)

	queryParams := r.URL.Query()
	page, err := strconv.Atoi(queryParams.Get("page"))

	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(queryParams.Get("limit"))

	if err != nil || limit < 1 {
		limit = h.cfg.DefaultPageLimit
	}

	params := service.GetOrdersParams{
		Page:  page,
		Limit: limit,
	}

	ordersList, total, err := h.orderService.GetOrders(ctx, params)
	if err != nil {
		log.Error(ctx, "Failed to get orders", zap.Error(err))
		h.writeErrorResponse(ctx, w, http.StatusInternalServerError, "Internal server error")
		return
	}

	ordersDTO := make([]dto.OrderDTO, 0, len(ordersList))
	for _, o := range ordersList {
		ordersDTO = append(ordersDTO, dto.OrderToDTO(o))
	}

	resp := &dto.OrdersResponse{
		Orders: ordersDTO,
		Total:  total,
		Page:   page,
	}

	h.writeJSONResponse(ctx, w, http.StatusOK, resp)
}
