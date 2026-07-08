package config

import (
	"fmt"

	"github.com/horizoonn/factory-platform/inventory/internal/config/env"
	platformenv "github.com/horizoonn/factory-platform/platform/pkg/config/env"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type envConfig struct {
	inventoryGRPC InventoryGRPCConfig
	migrations    MigrationsConfig
	app           AppConfig
	logger        logger.Config
}

func NewConfig() (Config, error) {
	inventoryGRPCConfig, err := env.NewInventoryGRPCConfig()
	if err != nil {
		return nil, fmt.Errorf("get inventory grpc config: %w", err)
	}

	migrationsConfig, err := env.NewMigrationsConfig()
	if err != nil {
		return nil, fmt.Errorf("get migrations config: %w", err)
	}

	appConfig, err := env.NewAppConfig()
	if err != nil {
		return nil, fmt.Errorf("get app config: %w", err)
	}

	loggerConfig, err := platformenv.NewLoggerConfig("inventory")
	if err != nil {
		return nil, fmt.Errorf("get logger config: %w", err)
	}

	return envConfig{
		inventoryGRPC: inventoryGRPCConfig,
		migrations:    migrationsConfig,
		app:           appConfig,
		logger:        loggerConfig,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get inventory config: %w", err)
		panic(err)
	}

	return config
}

func (c envConfig) InventoryGRPC() InventoryGRPCConfig {
	return c.inventoryGRPC
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
