package env

import (
	"fmt"
	"net"

	"github.com/kelseyhightower/envconfig"
)

type paymentGRPCConfig struct {
	Host string `envconfig:"HOST" default:"0.0.0.0"`
	Port string `envconfig:"PORT" required:"true"`
}

func NewPaymentGRPCConfig() (paymentGRPCConfig, error) {
	var config paymentGRPCConfig

	if err := envconfig.Process("PAYMENT_GRPC", &config); err != nil {
		return paymentGRPCConfig{}, fmt.Errorf("process payment grpc envconfig: %w", err)
	}

	return config, nil
}

func (c paymentGRPCConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
