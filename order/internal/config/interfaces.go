package config

import (
	"time"

	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type Config interface {
	OrderHTTP() OrderHTTPConfig
	InventoryGRPC() InventoryGRPCConfig
	PaymentGRPC() PaymentGRPCConfig
	Migrations() MigrationsConfig
	App() AppConfig
	Logger() logger.Config
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
