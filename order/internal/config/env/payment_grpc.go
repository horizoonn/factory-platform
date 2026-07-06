package env

import (
	"fmt"
	"net"

	"github.com/kelseyhightower/envconfig"
)

type PaymentGRPCConfig struct {
	Host string `envconfig:"HOST" required:"true"`
	Port string `envconfig:"PORT" required:"true"`
}

func NewPaymentGRPCConfig() (PaymentGRPCConfig, error) {
	var config PaymentGRPCConfig

	if err := envconfig.Process("PAYMENT_GRPC", &config); err != nil {
		return PaymentGRPCConfig{}, fmt.Errorf("process payment grpc envconfig: %w", err)
	}

	return config, nil
}

func (c PaymentGRPCConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
