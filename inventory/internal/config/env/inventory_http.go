package env

import (
	"fmt"
	"net"

	"github.com/kelseyhightower/envconfig"
)

type inventoryHTTPConfig struct {
	Host string `envconfig:"HOST" default:"0.0.0.0"`
	Port string `envconfig:"PORT" required:"true"`
}

func NewInventoryHTTPConfig() (inventoryHTTPConfig, error) {
	var config inventoryHTTPConfig

	if err := envconfig.Process("INVENTORY_HTTP", &config); err != nil {
		return inventoryHTTPConfig{}, fmt.Errorf("process inventory http envconfig: %w", err)
	}

	return config, nil
}

func (c inventoryHTTPConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
