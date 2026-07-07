package env

import (
	"fmt"
	"net"

	"github.com/kelseyhightower/envconfig"
)

type inventoryGRPCConfig struct {
	Host string `envconfig:"HOST" default:"0.0.0.0"`
	Port string `envconfig:"PORT" required:"true"`
}

func NewInventoryGRPCConfig() (inventoryGRPCConfig, error) {
	var config inventoryGRPCConfig

	if err := envconfig.Process("INVENTORY_GRPC", &config); err != nil {
		return inventoryGRPCConfig{}, fmt.Errorf("process inventory grpc envconfig: %w", err)
	}

	return config, nil
}

func (c inventoryGRPCConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
