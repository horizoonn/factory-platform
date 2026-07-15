package config

import (
	"time"

	outboxdispatcher "github.com/horizoonn/factory-platform/assembly/internal/outbox/dispatcher"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	consumerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer/franz"
	producerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/producer/franz"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type Config interface {
	Migrations() MigrationsConfig
	App() AppConfig
	Logger() logger.Config
	Postgres() pgxpool.Config
	OrderPaidConsumer() consumerfranz.Config
	ShipAssembledProducer() producerfranz.Config
	ShipAssembledTopic() string
	OutboxDispatcher() outboxdispatcher.Config
}

type MigrationsConfig interface {
	Dir() string
}

type AppConfig interface {
	ShutdownTimeout() time.Duration
}
