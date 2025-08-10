// main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"wb_tech_level_zero/internal/config"
	"wb_tech_level_zero/internal/gateway"
	"wb_tech_level_zero/pkg/db"

	"wb_tech_level_zero/internal/orders"

	appLogger "wb_tech_level_zero/pkg/logger"
)

const (
	serviceName = "wb_tech"
)

func main() {

	// Base logger
	bootstrapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer bootstrapLogger.Sync()

	if err := godotenv.Load(); err != nil {
		bootstrapLogger.Warn(".env file not found or cannot be read, relying on environment variables")
	}

	cfg := config.New()
	if cfg == nil {
		bootstrapLogger.Fatal("failed to load application configuration")
	}
	bootstrapLogger.Info("Configuration loaded")

	// Main context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgCfg := db.PostgresConfig{
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPassword,
		DBName:   cfg.PostgresDB,
	}
	pgPool, err := db.NewPostgresPool(ctx, pgCfg)
	if err != nil {
		bootstrapLogger.Fatal("failed to connect to Postgres", zap.Error(err))
	}
	defer pgPool.Close()
	bootstrapLogger.Info("Database connection pool established")

	// Orders
	ordersRepository := orders.NewPgOrdersRepository(pgPool)
	ordersService := orders.NewOrdersService(ordersRepository, cfg)
	ordersHandler := orders.NewOrdersHandler(ordersService, cfg)
	bootstrapLogger.Info("Orders services created")

	// Customs logger(adding serviceName and requestID in log)
	requestLogger := appLogger.New(bootstrapLogger, serviceName)
	ctx = context.WithValue(ctx, appLogger.LoggerKey, requestLogger)

	// GATEWAY
	gtw, err := gateway.New(ctx, cfg, ordersHandler)
	if err != nil {
		bootstrapLogger.Fatal("failed to init gateway", zap.Error(err))
	}

	// APP start
	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	runLogger := appLogger.GetLoggerFromCtx(ctx)

	go func() {
		runLogger.Info(ctx, "Starting server...", zap.String("address", cfg.HTTPServerAddress+":"+strconv.Itoa(cfg.HTTPServerPort)))
		if err := gtw.Run(ctx); err != nil {
			runLogger.Error(ctx, "gateway server stopped with error", zap.Error(err))
		}
		cancel()
	}()

	<-graceCh
	runLogger.Info(ctx, "Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := gtw.Shutdown(shutdownCtx); err != nil {
		runLogger.Error(ctx, "failed to shutdown gateway gracefully", zap.Error(err))
	} else {
		runLogger.Info(ctx, "Shutdown completed.")
	}
}
