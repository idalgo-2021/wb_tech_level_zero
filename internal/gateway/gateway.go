///////////////////////////////
// internal/gateway/gateway.go

package gateway

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"strconv"

	"wb_tech_level_zero/internal/config"
	httpapi "wb_tech_level_zero/internal/delivery/http"
	"wb_tech_level_zero/pkg/logger"

	"go.uber.org/zap"
)

type Server struct {
	cfg           *config.Config
	httpServer    *http.Server
	ordersHandler *httpapi.Handlers
}

func NewServer(ctx context.Context, cfg *config.Config, orderService httpapi.OrdersService) (*Server, error) {

	ordersHandler := httpapi.NewHandlers(cfg, orderService)

	r := NewRouter(ctx, ordersHandler)

	httpServer := &http.Server{
		Addr:    cfg.HTTPServerAddress + ":" + strconv.Itoa(cfg.HTTPServerPort),
		Handler: r,
	}

	httpServer.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	return &Server{
		httpServer:    httpServer,
		ordersHandler: ordersHandler,
		cfg:           cfg,
	}, nil
}

func (gtw *Server) Run(ctx context.Context) error {
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

func (g *Server) Shutdown(ctx context.Context) error {
	log := logger.GetLoggerFromCtx(ctx)
	log.Info(ctx, "HTTP server shutdown process initiated")
	if g.httpServer != nil {
		return g.httpServer.Shutdown(ctx)
	}
	return nil
}
