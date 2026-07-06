package config

import (
	"fmt"

	"github.com/horizoonn/factory-platform/payment/internal/config/env"
)

type EnvConfig struct {
	paymentGRPC env.PaymentGRPCConfig
	app         env.AppConfig
}

func NewConfig() (EnvConfig, error) {
	paymentGRPCConfig, err := env.NewPaymentGRPCConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get payment grpc config: %w", err)
	}

	appConfig, err := env.NewAppConfig()
	if err != nil {
		return EnvConfig{}, fmt.Errorf("get app config: %w", err)
	}

	return EnvConfig{
		paymentGRPC: paymentGRPCConfig,
		app:         appConfig,
	}, nil
}

func NewConfigMust() EnvConfig {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get payment config: %w", err)
		panic(err)
	}

	return config
}

func (c EnvConfig) PaymentGRPC() PaymentGRPCConfig {
	return c.paymentGRPC
}

func (c EnvConfig) App() AppConfig {
	return c.app
}
