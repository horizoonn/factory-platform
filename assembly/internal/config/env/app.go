package env

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type appConfig struct {
	ShutdownTimeoutValue time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"10s"`
}

func NewAppConfig() (appConfig, error) {
	var config appConfig

	if err := envconfig.Process("APP", &config); err != nil {
		return appConfig{}, fmt.Errorf("process app envconfig: %w", err)
	}

	return config, nil
}

func (c appConfig) ShutdownTimeout() time.Duration {
	return c.ShutdownTimeoutValue
}
