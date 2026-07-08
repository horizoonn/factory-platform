package config

import (
	"fmt"

	"github.com/horizoonn/factory-platform/inventory/internal/config/env"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type envConfig struct {
	inventoryGRPC InventoryGRPCConfig
	migrations    MigrationsConfig
	app           AppConfig
	logger        logger.Config
	postgres      pgxpool.Config
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

	loggerConfig, err := logger.NewConfigFromEnv("inventory")
	if err != nil {
		return nil, fmt.Errorf("get logger config: %w", err)
	}

	postgresConfig, err := pgxpool.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("get postgres config: %w", err)
	}

	return envConfig{
		inventoryGRPC: inventoryGRPCConfig,
		migrations:    migrationsConfig,
		app:           appConfig,
		logger:        loggerConfig,
		postgres:      postgresConfig,
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

func (c envConfig) Postgres() pgxpool.Config {
	return c.postgres
}
