package kafkadelivery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"wb_tech_level_zero/pkg/logger"

	kafkaGo "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type KafkaConfig struct {
	Brokers      []string
	GroupID      string
	Topic        string
	ConsumerCnt  int
	MaxRetries   int
	RetryDelayMs int
	TopicDLQ     string
}

type Consumer struct {
	reader       *kafkaGo.Reader
	handler      MessageHandler
	logger       logger.Logger
	wg           sync.WaitGroup
	closed       bool
	consumerCnt  int
	MaxRetries   int
	RetryDelayMs int
	TopicDLQ     string
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
		reader:       r,
		handler:      handler,
		logger:       logger,
		consumerCnt:  cfg.ConsumerCnt,
		MaxRetries:   cfg.MaxRetries,
		RetryDelayMs: cfg.RetryDelayMs,
		TopicDLQ:     cfg.TopicDLQ,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	for i := 0; i < c.consumerCnt; i++ {
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

				// msg processing
				if err := c.processMessageWithRetry(ctx, msg); err != nil {
					continue
				}

				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					c.logger.Error(ctx, fmt.Sprintf("Worker %d failed to commit Kafka message", workerID), zap.Error(err))
				}
			}
		}(i)
	}
	return nil
}

func (c *Consumer) processMessageWithRetry(ctx context.Context, msg kafkaGo.Message) error {
	var err error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {

		err = c.handler.HandleMessage(ctx, msg)
		if err == nil {
			return c.commit(ctx, msg)
		}

		if errors.Is(err, ErrKafkaNonRetryable) {
			c.logger.Warn(ctx, "Non-retryable error, sending to DLQ", zap.Error(err))
			c.sendToDLQAndCommit(ctx, msg)
			return nil
		}

		// ALL Not ErrKafkaNonRetryable errors

		if attempt < c.MaxRetries {
			delay := c.RetryDelayMs * (attempt + 1)
			c.logger.Warn(ctx, fmt.Sprintf("Retry %d for message, sleeping %d ms", attempt+1, delay), zap.Error(err))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(delay) * time.Millisecond):
				continue
			}
		}
	}

	c.logger.Error(ctx, "Max retries exceeded, sending to DLQ", zap.Error(err))
	c.sendToDLQAndCommit(ctx, msg)

	return err
}

func (c *Consumer) commit(ctx context.Context, msg kafkaGo.Message) error {
	if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
		c.logger.Error(ctx, "Failed to commit message", zap.Error(commitErr))
		return commitErr
	}
	return nil
}

func (c *Consumer) sendToDLQAndCommit(ctx context.Context, msg kafkaGo.Message) {
	if dlqErr := c.sendToDLQ(ctx, msg); dlqErr != nil {
		c.logger.Error(ctx, "Failed to send message to DLQ", zap.Error(dlqErr))
	}
	if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
		c.logger.Error(ctx, "Failed to commit message after DLQ", zap.Error(commitErr))
	}
}

func (c *Consumer) sendToDLQ(ctx context.Context, msg kafkaGo.Message) error {
	writer := kafkaGo.NewWriter(kafkaGo.WriterConfig{
		Brokers: c.reader.Config().Brokers,
		Topic:   c.TopicDLQ,
	})
	defer writer.Close()

	return writer.WriteMessages(ctx, kafkaGo.Message{
		Key:   msg.Key,
		Value: msg.Value,
	})
}

func (c *Consumer) Close() error {
	c.closed = true
	err := c.reader.Close()
	c.wg.Wait()
	return err
}
