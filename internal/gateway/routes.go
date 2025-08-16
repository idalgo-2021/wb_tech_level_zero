package gateway

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	httpapi "wb_tech_level_zero/internal/delivery/http"
	"wb_tech_level_zero/pkg/logger"

	"go.uber.org/zap"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(ctx context.Context, ordersHandler *httpapi.Handlers) *mux.Router {

	r := mux.NewRouter()
	r.Use(requestContextMiddleware)
	r.Use(corsMiddleware)

	// - - - - ORDERS
	r.HandleFunc("/order/{order_uid}", ordersHandler.GetOrderByUID).Methods(http.MethodGet)
	r.HandleFunc("/orders", ordersHandler.GetOrders).Methods(http.MethodGet)

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Генерируем уникальный ID для этого запроса.
		requestID := uuid.New().String()

		// 2. Создаем новый контекст, который содержит ТОЛЬКО что сгенерированный requestID.
		// Логгер уже лежит в r.Context() из main.go.
		ctxWithID := context.WithValue(r.Context(), logger.RequestIDKey, requestID)

		// 3. Логируем входящий запрос, используя новый контекст.
		// GetLoggerFromCtx извлечет базовый логгер из контекста,
		// а метод .Info() извлечет requestID из того же контекста.
		// Все работает автоматически!
		log := logger.GetLoggerFromCtx(ctxWithID)
		log.Info(ctxWithID, "Incoming request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
		)

		next.ServeHTTP(w, r.WithContext(ctxWithID))
	})
}
