package config

import (
	"time"

	outboxdispatcher "github.com/horizoonn/factory-platform/order/internal/outbox/dispatcher"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	consumerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer/franz"
	producerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/producer/franz"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type Config interface {
	OrderHTTP() OrderHTTPConfig
	InventoryGRPC() InventoryGRPCConfig
	PaymentGRPC() PaymentGRPCConfig
	Migrations() MigrationsConfig
	App() AppConfig
	Logger() logger.Config
	Postgres() pgxpool.Config
	OrderPaidProducer() producerfranz.Config
	OrderPaidTopic() string
	ShipAssembledConsumer() consumerfranz.Config
	OutboxDispatcher() outboxdispatcher.Config
}

type OrderHTTPConfig interface {
	Address() string
}

type InventoryGRPCConfig interface {
	Address() string
}

type PaymentGRPCConfig interface {
	Address() string
}

type MigrationsConfig interface {
	Dir() string
}

type AppConfig interface {
	ShutdownTimeout() time.Duration
}
