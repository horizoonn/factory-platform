package config

import (
	"time"

	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type Config interface {
	InventoryGRPC() InventoryGRPCConfig
	Migrations() MigrationsConfig
	App() AppConfig
	Logger() logger.Config
}

type InventoryGRPCConfig interface {
	Address() string
}

type MigrationsConfig interface {
	Dir() string
}

type AppConfig interface {
	ShutdownTimeout() time.Duration
}
