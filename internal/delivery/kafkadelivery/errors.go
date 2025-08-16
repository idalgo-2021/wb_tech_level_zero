package kafkadelivery

import "errors"

var (
	ErrKafkaRetryable    = errors.New("retryable error")
	ErrKafkaNonRetryable = errors.New("non-retryable error")
)
