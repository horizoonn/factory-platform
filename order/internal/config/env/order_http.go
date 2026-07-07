package env

import (
	"fmt"
	"net"

	"github.com/kelseyhightower/envconfig"
)

type orderHTTPConfig struct {
	Host string `envconfig:"HOST" default:"0.0.0.0"`
	Port string `envconfig:"PORT" required:"true"`
}

func NewOrderHTTPConfig() (orderHTTPConfig, error) {
	var config orderHTTPConfig

	if err := envconfig.Process("ORDER_HTTP", &config); err != nil {
		return orderHTTPConfig{}, fmt.Errorf("process order http envconfig: %w", err)
	}

	return config, nil
}

func (c orderHTTPConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
