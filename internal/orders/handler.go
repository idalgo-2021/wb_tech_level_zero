package orders

import (
	"errors"
	"net/http"
	"strconv"
	"wb_tech_level_zero/internal/config"
	"wb_tech_level_zero/internal/models"
	"wb_tech_level_zero/pkg/logger"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type OrdersHandler struct {
	service *OrdersService
	cfg     *config.Config
}

func NewOrdersHandler(service *OrdersService, cfg *config.Config) *OrdersHandler {
	return &OrdersHandler{service: service, cfg: cfg}
}

func (h *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
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

	dto := GetOrdersRequest{
		Page:  page,
		Limit: limit,
	}

	orders, err := h.service.GetOrders(ctx, dto)
	if err != nil {
		log.Error(ctx, "Failed to get orders", zap.Error(err))
		h.writeErrorResponse(ctx, w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.writeJSONResponse(ctx, w, http.StatusOK, orders)

}

func (h *OrdersHandler) GetOrderByUID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLoggerFromCtx(ctx)

	vars := mux.Vars(r)
	orderUID, ok := vars["order_uid"]
	if !ok || orderUID == "" {
		log.Error(ctx, "URL path parameter 'order_uid' is missing")
		h.writeErrorResponse(ctx, w, http.StatusBadRequest, "order_uid path parameter is required")
		return
	}

	order, err := h.service.GetOrderByUID(ctx, orderUID)
	if err != nil {
		if errors.Is(err, models.ErrOrderNotFound) {
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

	h.writeJSONResponse(ctx, w, http.StatusOK, order)
}

// func (h *OrdersHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	log := logger.GetLoggerFromCtx(ctx)
//
// 	orderID, ok := h.parseUUIDFromPath(w, r, "id")
// 	if !ok {
// 		return
// 	}
//
// 	order, err := h.service.GetOrderByID(ctx, orderID)
// 	if err != nil {
// 		if errors.Is(err, models.ErrOrderNotFound) {
// 			log.Info(ctx, "Order not found by ID", zap.String("order_id", orderID.String()))
// 			h.writeErrorResponse(ctx, w, http.StatusNotFound, "Order not found")
// 		} else {
// 			log.Error(ctx, "Failed to get order by ID",
// 				zap.Error(err),
// 				zap.String("order_id", orderID.String()),
// 			)
// 			h.writeErrorResponse(ctx, w, http.StatusInternalServerError, "Internal server error")
// 		}
// 		return
// 	}
//
// 	h.writeJSONResponse(ctx, w, http.StatusOK, order)
//
// }
