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

	KafkaBroker  string `env:"KAFKA_BROKER" envDefault:"localhost:9094"`
	KafkaTopic   string `env:"KAFKA_TOPIC" envDefault:"orders"`
	KafkaGroupID string `env:"KAFKA_GROUP_ID" envDefault:"order-consumer"`
}

func New() *Config {
	cfg := Config{}
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil
	}
	return &cfg
}
