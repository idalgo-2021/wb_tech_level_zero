package gateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"wb_tech_level_zero/internal/config"
	"wb_tech_level_zero/internal/orders"
	"wb_tech_level_zero/pkg/logger"

	"go.uber.org/zap"
)

type Gateway struct {
	httpServer *http.Server
}

func New(
	ctx context.Context,
	cfg *config.Config,
	ordersHandler *orders.OrdersHandler,
) (*Gateway, error) {

	r := NewRouter(ctx, ordersHandler)

	httpServer := &http.Server{
		Addr:    cfg.HTTPServerAddress + ":" + strconv.Itoa(cfg.HTTPServerPort),
		Handler: r,
	}

	httpServer.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	return &Gateway{
		httpServer: httpServer,
	}, nil
}

func (gtw *Gateway) Run(ctx context.Context) error {
	log := logger.GetLoggerFromCtx(ctx)
	log.Info(ctx, "HTTP server is starting to listen")

	err := gtw.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Error(ctx,
			"HTTP server ListenAndServe failed",
			zap.Error(err),
		)
		return fmt.Errorf("http server ListenAndServe error: %w", err)
	}
	return nil
}

func (g *Gateway) Shutdown(ctx context.Context) error {
	log := logger.GetLoggerFromCtx(ctx)
	log.Info(ctx, "HTTP server shutdown process initiated")
	if g.httpServer != nil {
		return g.httpServer.Shutdown(ctx)
	}
	return nil
}
