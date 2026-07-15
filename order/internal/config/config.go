package config

import (
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/config/env"
	outboxdispatcher "github.com/horizoonn/factory-platform/order/internal/outbox/dispatcher"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	consumerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer/franz"
	producerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/producer/franz"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type envConfig struct {
	orderHTTP             OrderHTTPConfig
	inventoryGRPC         InventoryGRPCConfig
	paymentGRPC           PaymentGRPCConfig
	migrations            MigrationsConfig
	app                   AppConfig
	logger                logger.Config
	postgres              pgxpool.Config
	orderPaidProducer     producerfranz.Config
	orderPaidTopic        string
	shipAssembledConsumer consumerfranz.Config
	outboxDispatcher      outboxdispatcher.Config
}

func NewConfig() (Config, error) {
	orderHTTPConfig, err := env.NewOrderHTTPConfig()
	if err != nil {
		return nil, fmt.Errorf("get order http config: %w", err)
	}

	inventoryGRPCConfig, err := env.NewInventoryGRPCConfig()
	if err != nil {
		return nil, fmt.Errorf("get inventory grpc config: %w", err)
	}

	paymentGRPCConfig, err := env.NewPaymentGRPCConfig()
	if err != nil {
		return nil, fmt.Errorf("get payment grpc config: %w", err)
	}

	migrationsConfig, err := env.NewMigrationsConfig()
	if err != nil {
		return nil, fmt.Errorf("get migrations config: %w", err)
	}

	appConfig, err := env.NewAppConfig()
	if err != nil {
		return nil, fmt.Errorf("get app config: %w", err)
	}

	loggerConfig, err := logger.NewConfigFromEnv("order")
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

	orderPaidProducerConfig, orderPaidTopic, err := env.NewOrderPaidProducerConfig(brokers)
	if err != nil {
		return nil, fmt.Errorf("get order paid producer config: %w", err)
	}

	shipAssembledConsumerConfig, err := env.NewShipAssembledConsumerConfig(brokers)
	if err != nil {
		return nil, fmt.Errorf("get ship assembled consumer config: %w", err)
	}

	outboxDispatcherConfig, err := env.NewOutboxDispatcherConfig()
	if err != nil {
		return nil, fmt.Errorf("get outbox dispatcher config: %w", err)
	}
	if orderPaidProducerConfig.DeliveryTimeout > outboxDispatcherConfig.PublishTimeout {
		return nil, fmt.Errorf(
			"order paid producer delivery timeout %s exceeds outbox publish timeout %s",
			orderPaidProducerConfig.DeliveryTimeout,
			outboxDispatcherConfig.PublishTimeout,
		)
	}

	return envConfig{
		orderHTTP:             orderHTTPConfig,
		inventoryGRPC:         inventoryGRPCConfig,
		paymentGRPC:           paymentGRPCConfig,
		migrations:            migrationsConfig,
		app:                   appConfig,
		logger:                loggerConfig,
		postgres:              postgresConfig,
		orderPaidProducer:     orderPaidProducerConfig,
		orderPaidTopic:        orderPaidTopic,
		shipAssembledConsumer: shipAssembledConsumerConfig,
		outboxDispatcher:      outboxDispatcherConfig,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get order config: %w", err)
		panic(err)
	}

	return config
}

func (c envConfig) OrderHTTP() OrderHTTPConfig {
	return c.orderHTTP
}

func (c envConfig) InventoryGRPC() InventoryGRPCConfig {
	return c.inventoryGRPC
}

func (c envConfig) PaymentGRPC() PaymentGRPCConfig {
	return c.paymentGRPC
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

func (c envConfig) OrderPaidProducer() producerfranz.Config {
	return c.orderPaidProducer
}

func (c envConfig) OrderPaidTopic() string {
	return c.orderPaidTopic
}

func (c envConfig) ShipAssembledConsumer() consumerfranz.Config {
	return c.shipAssembledConsumer
}

func (c envConfig) OutboxDispatcher() outboxdispatcher.Config {
	return c.outboxDispatcher
}
