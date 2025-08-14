package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPServerAddress string `env:"HTTP_SERVER_ADDRESS" env-default:"localhost"`
	HTTPServerPort    int    `env:"HTTP_SERVER_PORT" env-default:"8080"`

	PostgresHost     string `env:"POSTGRES_HOST" env-default:"localhost"`
	PostgresPort     int    `env:"POSTGRES_PORT" env-default:"6432"`
	PostgresUser     string `env:"POSTGRES_USER" env-default:"pguser"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" env-default:"pgpass"`
	PostgresDB       string `env:"POSTGRES_DB" env-default:"wbdb"`

	DefaultPageLimit int `env:"DEFAULT_PAGE_LIMIT" env-default:"50"`

	KafkaBroker        string `env:"KAFKA_BROKER" env-default:"localhost:9094"`
	KafkaTopic         string `env:"KAFKA_TOPIC" env-default:"orders"`
	KafkaGroupID       string `env:"KAFKA_GROUP_ID" env-default:"order-consumer"`
	KafkaConsumerCount int    `env:"KAFKA_CONSUMER_COUNT" env-default:"2"`

	KafkaMaxRetries   int    `env:"KAFKA_MAX_RETRIES" env-default:"3"`
	KafkaRetryDelayMs int    `env:"KAFKA_RETRY_DELAY_MS" env-default:"600"`
	KafkaTopicDLQ     string `env:"KAFKA_DLQ_TOPIC" env-default:"orders-dlq"`
}

func New() (*Config, error) {
	cfg := Config{}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
