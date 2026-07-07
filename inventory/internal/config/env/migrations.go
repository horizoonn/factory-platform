package env

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type MigrationsConfig struct {
	DirValue string `envconfig:"DIR" required:"true"`
}

func NewMigrationsConfig() (MigrationsConfig, error) {
	var config MigrationsConfig

	if err := envconfig.Process("INVENTORY_MIGRATIONS", &config); err != nil {
		return MigrationsConfig{}, fmt.Errorf("process inventory migrations envconfig: %w", err)
	}

	return config, nil
}

func (c MigrationsConfig) Dir() string {
	return c.DirValue
}
