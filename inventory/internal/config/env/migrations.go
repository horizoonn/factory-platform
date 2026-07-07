package env

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type migrationsConfig struct {
	DirValue string `envconfig:"DIR" required:"true"`
}

func NewMigrationsConfig() (migrationsConfig, error) {
	var config migrationsConfig

	if err := envconfig.Process("INVENTORY_MIGRATIONS", &config); err != nil {
		return migrationsConfig{}, fmt.Errorf("process inventory migrations envconfig: %w", err)
	}

	return config, nil
}

func (c migrationsConfig) Dir() string {
	return c.DirValue
}
