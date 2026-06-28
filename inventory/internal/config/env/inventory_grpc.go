package env

import (
	"fmt"
	"net"

	"github.com/kelseyhightower/envconfig"
)

type InventoryGRPCConfig struct {
	Host string `envconfig:"HOST" default:"0.0.0.0"`
	Port string `envconfig:"PORT" required:"true"`
}

func NewInventoryGRPCConfig() (InventoryGRPCConfig, error) {
	var config InventoryGRPCConfig

	if err := envconfig.Process("INVENTORY_GRPC", &config); err != nil {
		return InventoryGRPCConfig{}, fmt.Errorf("process inventory grpc envconfig: %w", err)
	}

	return config, nil
}

func (c InventoryGRPCConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
