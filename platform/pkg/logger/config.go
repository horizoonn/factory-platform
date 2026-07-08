package logger

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

const defaultLevel = "info"

type Config struct {
	Level       string
	JSON        bool
	Development bool
	ServiceName string
}

type envConfig struct {
	LevelValue  string `envconfig:"LEVEL" default:"info"`
	AsJsonValue bool   `envconfig:"AS_JSON" default:"true"`
	Development bool   `envconfig:"DEVELOPMENT" default:"false"`
}

func NewConfigFromEnv(serviceName string) (Config, error) {
	var raw envConfig
	if err := envconfig.Process("LOGGER", &raw); err != nil {
		return Config{}, fmt.Errorf("process logger envconfig: %w", err)
	}

	return Config{
		Level:       raw.LevelValue,
		JSON:        raw.AsJsonValue,
		Development: raw.Development,
		ServiceName: serviceName,
	}, nil
}
