package env

import (
	"errors"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"

	producerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/producer/franz"
)

type shipAssembledProducerConfig struct {
	ClientID        string        `envconfig:"CLIENT_ID" default:"assembly-ship-assembled-producer"`
	Topic           string        `envconfig:"TOPIC" required:"true"`
	DeliveryTimeout time.Duration `envconfig:"DELIVERY_TIMEOUT" default:"10s"`
}

func NewShipAssembledProducerConfig(brokers []string) (producerfranz.Config, string, error) {
	var config shipAssembledProducerConfig

	if err := envconfig.Process("SHIP_ASSEMBLED_PRODUCER", &config); err != nil {
		return producerfranz.Config{}, "", fmt.Errorf("process ship assembled producer envconfig: %w", err)
	}
	if config.DeliveryTimeout <= 0 {
		return producerfranz.Config{}, "", errors.New("ship assembled producer delivery timeout must be positive")
	}

	return producerfranz.Config{
		Brokers:         brokers,
		ClientID:        config.ClientID,
		DeliveryTimeout: config.DeliveryTimeout,
	}, config.Topic, nil
}
