package config

import (
	"fmt"

	"github.com/horizoonn/factory-platform/payment/internal/config/env"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type envConfig struct {
	paymentGRPC PaymentGRPCConfig
	app         AppConfig
	logger      logger.Config
}

func NewConfig() (Config, error) {
	paymentGRPCConfig, err := env.NewPaymentGRPCConfig()
	if err != nil {
		return nil, fmt.Errorf("get payment grpc config: %w", err)
	}

	appConfig, err := env.NewAppConfig()
	if err != nil {
		return nil, fmt.Errorf("get app config: %w", err)
	}

	loggerConfig, err := logger.NewConfigFromEnv("payment")
	if err != nil {
		return nil, fmt.Errorf("get logger config: %w", err)
	}

	return envConfig{
		paymentGRPC: paymentGRPCConfig,
		app:         appConfig,
		logger:      loggerConfig,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get payment config: %w", err)
		panic(err)
	}

	return config
}

func (c envConfig) PaymentGRPC() PaymentGRPCConfig {
	return c.paymentGRPC
}

func (c envConfig) App() AppConfig {
	return c.app
}

func (c envConfig) Logger() logger.Config {
	return c.logger
}
