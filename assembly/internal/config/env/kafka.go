package env

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type kafkaConfig struct {
	Brokers []string `envconfig:"BROKERS" required:"true"`
}

func NewKafkaBrokers() ([]string, error) {
	var config kafkaConfig

	if err := envconfig.Process("KAFKA", &config); err != nil {
		return nil, fmt.Errorf("process kafka envconfig: %w", err)
	}

	return config.Brokers, nil
}
