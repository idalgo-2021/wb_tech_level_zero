///////////////////////

// cmd/main/main.go
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"wb_tech_level_zero/internal/app"
	"wb_tech_level_zero/internal/config"
	appLogger "wb_tech_level_zero/pkg/logger"
)

const (
	serviceName = "wb_tech"
)

func main() {
	bootstrapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer bootstrapLogger.Sync()

	if err := godotenv.Load(); err != nil {
		bootstrapLogger.Warn(".env file not found or cannot be read, relying on environment variables")
	}

	cfg, err := config.New()
	if err != nil {
		bootstrapLogger.Fatal("failed to load application configuration", zap.Error(err))
	}
	bootstrapLogger.Info("Configuration loaded")

	logger := appLogger.New(bootstrapLogger, serviceName)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg, logger)
	if err != nil {
		logger.Fatal(ctx, "failed to init app", zap.Error(err))
	}

	if err := application.Run(ctx); err != nil {
		logger.Fatal(ctx, "failed to start app", zap.Error(err))
	}

	<-ctx.Done()
	logger.Info(ctx, "Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Stop(shutdownCtx); err != nil {
		logger.Error(ctx, "shutdown error", zap.Error(err))
	}
}
