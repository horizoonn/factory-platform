package env

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	ShutdownTimeoutValue time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"10s"`
}

func NewAppConfig() (AppConfig, error) {
	var config AppConfig

	if err := envconfig.Process("APP", &config); err != nil {
		return AppConfig{}, fmt.Errorf("process app envconfig: %w", err)
	}

	return config, nil
}

func (c AppConfig) ShutdownTimeout() time.Duration {
	return c.ShutdownTimeoutValue
}
