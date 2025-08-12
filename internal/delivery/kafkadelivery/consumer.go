//
// consumer.go

package kafkadelivery

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"wb_tech_level_zero/pkg/logger"

	kafkaGo "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaConfig struct {
	Brokers []string
	GroupID string
	Topic   string
}

type Consumer struct {
	reader  *kafkaGo.Reader
	handler MessageHandler
	logger  logger.Logger
	wg      sync.WaitGroup
	closed  bool
}

type MessageHandler interface {
	HandleMessage(ctx context.Context, msg kafkaGo.Message) error
}

func NewConsumer(cfg KafkaConfig, handler MessageHandler, logger logger.Logger) *Consumer {
	r := kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{
		reader:  r,
		handler: handler,
		logger:  logger,
	}
}

func (c *Consumer) Start(ctx context.Context, workers int) error {
	for i := 0; i < workers; i++ {
		c.wg.Add(1)
		go func(workerID int) {
			defer c.wg.Done()
			c.logger.Info(ctx, fmt.Sprintf("Worker %d started", workerID))
			for {
				msg, err := c.reader.FetchMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						c.logger.Error(ctx, fmt.Sprintf("Worker %d topped by context cancel", workerID), zap.Error(err))
						return
					}
					c.logger.Error(ctx, "Failed to fetch message from Kafka", zap.Error(err))
					continue
				}

				if err := c.handler.HandleMessage(ctx, msg); err != nil {
					c.logger.Error(ctx, fmt.Sprintf("Worker %d failed to handle message", workerID), zap.Error(err))
					// Retry, DLQ ....
				}

				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					c.logger.Error(ctx, fmt.Sprintf("Worker %d failed to commit Kafka message", workerID), zap.Error(err))
				}
			}
		}(i)
	}
	return nil
}

func (c *Consumer) Close() error {
	c.closed = true
	err := c.reader.Close()
	c.wg.Wait()
	return err
}
