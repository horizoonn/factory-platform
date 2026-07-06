package config

import (
	"fmt"

	"github.com/horizoonn/factory-platform/inventory/internal/config/env"
)

type EnvConfig struct {
	inventoryGRPC env.InventoryGRPCConfig
	app           env.AppConfig
}

func NewConfig() (EnvConfig, error) {
	inventoryGRPCConfig, err := env.NewInventoryGRPCConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get inventory grpc config: %w", err)
	}

	appConfig, err := env.NewAppConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get app config: %w", err)
	}

	return EnvConfig{
		inventoryGRPC: inventoryGRPCConfig,
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

func (c EnvConfig) App() AppConfig {
	return c.app
}
