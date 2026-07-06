package env

import (
	"fmt"
	"net"

	"github.com/kelseyhightower/envconfig"
)

type OrderHTTPConfig struct {
	Host string `envconfig:"HOST" default:"0.0.0.0"`
	Port string `envconfig:"PORT" required:"true"`
}

func NewOrderHTTPConfig() (OrderHTTPConfig, error) {
	var config OrderHTTPConfig

	if err := envconfig.Process("ORDER_HTTP", &config); err != nil {
		return OrderHTTPConfig{}, fmt.Errorf("process order http envconfig: %w", err)
	}

	return config, nil
}

func (c OrderHTTPConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
