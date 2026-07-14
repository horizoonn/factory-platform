package config

import (
	"fmt"

	"github.com/horizoonn/factory-platform/assembly/internal/config/env"
	outboxdispatcher "github.com/horizoonn/factory-platform/assembly/internal/outbox/dispatcher"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	consumerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer/franz"
	producerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/producer/franz"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type envConfig struct {
	migrations            MigrationsConfig
	app                   AppConfig
	logger                logger.Config
	postgres              pgxpool.Config
	orderPaidConsumer     consumerfranz.Config
	shipAssembledProducer producerfranz.Config
	outboxDispatcher      outboxdispatcher.Config
}

func NewConfig() (Config, error) {
	migrationsConfig, err := env.NewMigrationsConfig()
	if err != nil {
		return nil, fmt.Errorf("get migrations config: %w", err)
	}

	appConfig, err := env.NewAppConfig()
	if err != nil {
		return nil, fmt.Errorf("get app config: %w", err)
	}

	loggerConfig, err := logger.NewConfigFromEnv("assembly")
	if err != nil {
		return nil, fmt.Errorf("get logger config: %w", err)
	}

	postgresConfig, err := pgxpool.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("get postgres config: %w", err)
	}

	brokers, err := env.NewKafkaBrokers()
	if err != nil {
		return nil, fmt.Errorf("get kafka brokers: %w", err)
	}

	orderPaidConsumerConfig, err := env.NewOrderPaidConsumerConfig(brokers)
	if err != nil {
		return nil, fmt.Errorf("get order paid consumer config: %w", err)
	}

	shipAssembledProducerConfig, err := env.NewShipAssembledProducerConfig(brokers)
	if err != nil {
		return nil, fmt.Errorf("get ship assembled producer config: %w", err)
	}

	outboxDispatcherConfig, err := env.NewOutboxDispatcherConfig()
	if err != nil {
		return nil, fmt.Errorf("get outbox dispatcher config: %w", err)
	}
	if shipAssembledProducerConfig.DeliveryTimeout > outboxDispatcherConfig.PublishTimeout {
		return nil, fmt.Errorf(
			"ship assembled producer delivery timeout %s exceeds outbox publish timeout %s",
			shipAssembledProducerConfig.DeliveryTimeout,
			outboxDispatcherConfig.PublishTimeout,
		)
	}

	return envConfig{
		migrations:            migrationsConfig,
		app:                   appConfig,
		logger:                loggerConfig,
		postgres:              postgresConfig,
		orderPaidConsumer:     orderPaidConsumerConfig,
		shipAssembledProducer: shipAssembledProducerConfig,
		outboxDispatcher:      outboxDispatcherConfig,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get assembly config: %w", err)
		panic(err)
	}

	return config
}

func (c envConfig) Migrations() MigrationsConfig {
	return c.migrations
}

func (c envConfig) App() AppConfig {
	return c.app
}

func (c envConfig) Logger() logger.Config {
	return c.logger
}

func (c envConfig) Postgres() pgxpool.Config {
	return c.postgres
}

func (c envConfig) OrderPaidConsumer() consumerfranz.Config {
	return c.orderPaidConsumer
}

func (c envConfig) ShipAssembledProducer() producerfranz.Config {
	return c.shipAssembledProducer
}

func (c envConfig) OutboxDispatcher() outboxdispatcher.Config {
	return c.outboxDispatcher
}
