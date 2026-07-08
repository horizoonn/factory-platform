package config

import (
	"time"

	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type Config interface {
	PaymentGRPC() PaymentGRPCConfig
	App() AppConfig
	Logger() logger.Config
}

type PaymentGRPCConfig interface {
	Address() string
}

type AppConfig interface {
	ShutdownTimeout() time.Duration
}
