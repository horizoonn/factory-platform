package env

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type loggerConfig struct {
	LevelValue  string `envconfig:"LEVEL" default:"info"`
	AsJsonValue bool   `envconfig:"AS_JSON" default:"true"`
	Development bool   `envconfig:"DEVELOPMENT" default:"false"`
}

func NewLoggerConfig(serviceName string) (logger.Config, error) {
	var raw loggerConfig
	if err := envconfig.Process("LOGGER", &raw); err != nil {
		return logger.Config{}, fmt.Errorf("process logger envconfig: %w", err)
	}

	return logger.Config{
		Level:       raw.LevelValue,
		JSON:        raw.AsJsonValue,
		Development: raw.Development,
		ServiceName: serviceName,
	}, nil
}
