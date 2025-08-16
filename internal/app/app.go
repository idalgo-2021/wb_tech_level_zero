package app

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"wb_tech_level_zero/internal/config"
	"wb_tech_level_zero/internal/gateway"

	"wb_tech_level_zero/internal/cache"
	"wb_tech_level_zero/internal/delivery/kafkadelivery"
	"wb_tech_level_zero/internal/repository"
	"wb_tech_level_zero/internal/service"
	"wb_tech_level_zero/pkg/db"
	"wb_tech_level_zero/pkg/logger"

	"wb_tech_level_zero/pkg/redisclient"

	_ "wb_tech_level_zero/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type App struct {
	cfg           *config.Config
	logger        logger.Logger
	httpServer    *gateway.Server
	kafkaConsumer *kafkadelivery.Consumer
	pgPool        *pgxpool.Pool
	redisClient   *redis.Client
	orderService  service.OrdersService
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

	redisCfg := redisclient.RedisConfig{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	redisClient, err := redisclient.New(ctx, redisCfg)
	if err != nil {
		return nil, err
	}

	cacheCfg := cache.CacheConfig{
		TTL: cfg.OrderTTLMinutes,
	}
	orderCache := cache.NewOrdersCache(redisClient, cacheCfg)

	app := &App{
		cfg:         cfg,
		logger:      logger,
		pgPool:      pgPool,
		redisClient: redisClient,
	}

	app.orderService = service.NewOrdersService(cfg, orderRepo, orderCache, &app.wg, logger)

	app.httpServer, err = gateway.NewServer(ctx, cfg, app.orderService)
	if err != nil {
		logger.Fatal(ctx, "failed to init gateway", zap.Error(err))
		return nil, err
	}

	kafkaHandler := kafkadelivery.NewHandler(app.orderService, logger)
	kafkaCfg := kafkadelivery.KafkaConfig{
		Brokers:      strings.Split(cfg.KafkaBroker, ","),
		GroupID:      cfg.KafkaGroupID,
		Topic:        cfg.KafkaTopic,
		ConsumerCnt:  cfg.KafkaConsumerCount,
		MaxRetries:   cfg.KafkaMaxRetries,
		RetryDelayMs: cfg.KafkaRetryDelayMs,
		TopicDLQ:     cfg.KafkaTopicDLQ,
	}
	app.kafkaConsumer = kafkadelivery.NewConsumer(kafkaCfg, kafkaHandler, logger)

	return app, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx = logger.ContextWithLogger(ctx, a.logger)

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.logger.Info(ctx, "Warming up Redis order cache...")
		if err := a.orderService.WarmOrdersCache(ctx); err != nil {
			a.logger.Error(ctx, "Failed to warm up cache", zap.Error(err))
		}
	}()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.logger.Info(ctx, "Starting Kafka consumer...")
		if err := a.kafkaConsumer.Start(ctx); err != nil {
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

	a.logger.Info(ctx, "Closing Redis connection")
	if err := a.redisClient.Close(); err != nil {
		a.logger.Error(ctx, "Redis close error", zap.Error(err))
	}

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
