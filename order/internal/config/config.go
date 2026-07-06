package config

import (
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/config/env"
)

type EnvConfig struct {
	orderHTTP     env.OrderHTTPConfig
	inventoryGRPC env.InventoryGRPCConfig
	paymentGRPC   env.PaymentGRPCConfig
	app           env.AppConfig
}

func NewConfig() (EnvConfig, error) {
	orderHTTPConfig, err := env.NewOrderHTTPConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get order http config: %w", err)
	}

	inventoryGRPCConfig, err := env.NewInventoryGRPCConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get inventory grpc config: %w", err)
	}

	paymentGRPCConfig, err := env.NewPaymentGRPCConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get payment grpc config: %w", err)
	}

	appConfig, err := env.NewAppConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get app config: %w", err)
	}

	return EnvConfig{
		orderHTTP:     orderHTTPConfig,
		inventoryGRPC: inventoryGRPCConfig,
		paymentGRPC:   paymentGRPCConfig,
		app:           appConfig,
	}, nil
}

func NewConfigMust() EnvConfig {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get order config: %w", err)
		panic(err)
	}

	return config
}

func (c EnvConfig) OrderHTTP() OrderHTTPConfig {
	return c.orderHTTP
}

func (c EnvConfig) InventoryGRPC() InventoryGRPCConfig {
	return c.inventoryGRPC
}

func (c EnvConfig) PaymentGRPC() PaymentGRPCConfig {
	return c.paymentGRPC
}

func (c EnvConfig) App() AppConfig {
	return c.app
}
