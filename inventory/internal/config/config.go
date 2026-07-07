package config

import (
	"fmt"

	"github.com/horizoonn/factory-platform/inventory/internal/config/env"
)

type EnvConfig struct {
	inventoryGRPC env.InventoryGRPCConfig
	migrations    env.MigrationsConfig
	app           env.AppConfig
}

func NewConfig() (EnvConfig, error) {
	inventoryGRPCConfig, err := env.NewInventoryGRPCConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get inventory grpc config: %w", err)
	}

	migrationsConfig, err := env.NewMigrationsConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get migrations config: %w", err)
	}

	appConfig, err := env.NewAppConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get app config: %w", err)
	}

	return EnvConfig{
		inventoryGRPC: inventoryGRPCConfig,
		migrations:    migrationsConfig,
		app:           appConfig,
	}, nil
}

func NewConfigMust() EnvConfig {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get inventory config: %w", err)
		panic(err)
	}

	return config
}

func (c EnvConfig) InventoryGRPC() InventoryGRPCConfig {
	return c.inventoryGRPC
}

func (c EnvConfig) Migrations() MigrationsConfig {
	return c.migrations
}

func (c EnvConfig) App() AppConfig {
	return c.app
}
