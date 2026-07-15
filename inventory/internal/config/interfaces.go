package config

import (
	"time"

	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type Config interface {
	InventoryGRPC() InventoryGRPCConfig
	InventoryHTTP() InventoryHTTPConfig
	Migrations() MigrationsConfig
	App() AppConfig
	Logger() logger.Config
	Postgres() pgxpool.Config
}

type InventoryGRPCConfig interface {
	Address() string
}

type InventoryHTTPConfig interface {
	Address() string
}

type MigrationsConfig interface {
	Dir() string
}

type AppConfig interface {
	ShutdownTimeout() time.Duration
}
