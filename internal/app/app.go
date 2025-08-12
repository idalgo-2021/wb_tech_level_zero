///////////////////////

// internal/app/app.go

package app

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"wb_tech_level_zero/internal/config"
	"wb_tech_level_zero/internal/gateway"

	"wb_tech_level_zero/internal/delivery/kafkadelivery"
	"wb_tech_level_zero/internal/repository"
	"wb_tech_level_zero/internal/service"
	"wb_tech_level_zero/pkg/db"
	"wb_tech_level_zero/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	cfg           *config.Config
	logger        logger.Logger
	httpServer    *gateway.Server
	kafkaConsumer *kafkadelivery.Consumer
	pgPool        *pgxpool.Pool
	wg            sync.WaitGroup
}

func New(ctx context.Context, cfg *config.Config, logger logger.Logger) (*App, error) {

	pgCfg := db.PostgresConfig{
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPassword,
		DBName:   cfg.PostgresDB,
	}
	pgPool, err := db.NewPostgresPool(ctx, pgCfg)
	if err != nil {
		return nil, err
	}
	orderRepo := repository.NewOrdersRepository(pgPool)
	orderService := service.NewOrdersService(cfg, orderRepo)

	httpServer, err := gateway.NewServer(ctx, cfg, orderService)
	if err != nil {
		logger.Fatal(ctx, "failed to init gateway", zap.Error(err))
		return nil, err
	}

	kafkaHandler := kafkadelivery.NewHandler(orderService, logger)
	kafkaCfg := kafkadelivery.KafkaConfig{
		Brokers: strings.Split(cfg.KafkaBroker, ","),
		GroupID: cfg.KafkaGroupID,
		Topic:   cfg.KafkaTopic,
	}
	kafkaConsumer := kafkadelivery.NewConsumer(kafkaCfg, kafkaHandler, logger)

	return &App{
		cfg:           cfg,
		logger:        logger,
		pgPool:        pgPool,
		httpServer:    httpServer,
		kafkaConsumer: kafkaConsumer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.logger.Info(ctx, "Starting Kafka consumer...")
		if err := a.kafkaConsumer.Start(ctx, 3); err != nil {
			a.logger.Error(ctx, "Kafka consumer failed", zap.Error(err))
		}
	}()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.logger.Info(ctx, "Starting HTTP server on "+a.cfg.HTTPServerAddress+":"+strconv.Itoa(a.cfg.HTTPServerPort))
		if err := a.httpServer.Run(ctx); err != nil {
			a.logger.Error(ctx, "HTTP server stopped with error", zap.Error(err))
		}
	}()

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	a.logger.Info(ctx, "Stopping HTTP server")
	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.logger.Error(ctx, "HTTP server shutdown error", zap.Error(err))
	}

	a.logger.Info(ctx, "Stopping Kafka consumer")
	if err := a.kafkaConsumer.Close(); err != nil {
		a.logger.Error(ctx, "Kafka consumer shutdown error", zap.Error(err))
	}

	a.logger.Info(ctx, "Closing DB connection")
	a.pgPool.Close()

	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info(ctx, "All goroutines stopped")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
