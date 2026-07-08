package config

import (
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/config/env"
	platformenv "github.com/horizoonn/factory-platform/platform/pkg/config/env"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type envConfig struct {
	orderHTTP     OrderHTTPConfig
	inventoryGRPC InventoryGRPCConfig
	paymentGRPC   PaymentGRPCConfig
	migrations    MigrationsConfig
	app           AppConfig
	logger        logger.Config
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

	loggerConfig, err := platformenv.NewLoggerConfig("order")
	if err != nil {
		return nil, fmt.Errorf("get logger config: %w", err)
	}

	return envConfig{
		orderHTTP:     orderHTTPConfig,
		inventoryGRPC: inventoryGRPCConfig,
		paymentGRPC:   paymentGRPCConfig,
		migrations:    migrationsConfig,
		app:           appConfig,
		logger:        loggerConfig,
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
