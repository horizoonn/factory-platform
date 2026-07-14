package env

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"

	producerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/producer/franz"
)

type shipAssembledProducerConfig struct {
	ClientID        string        `envconfig:"CLIENT_ID" default:"assembly-ship-assembled-producer"`
	DeliveryTimeout time.Duration `envconfig:"DELIVERY_TIMEOUT" default:"10s"`
}

func NewShipAssembledProducerConfig(brokers []string) (producerfranz.Config, error) {
	var config shipAssembledProducerConfig

	if err := envconfig.Process("SHIP_ASSEMBLED_PRODUCER", &config); err != nil {
		return producerfranz.Config{}, fmt.Errorf("process ship assembled producer envconfig: %w", err)
	}

	return producerfranz.Config{
		Brokers:         brokers,
		ClientID:        config.ClientID,
		DeliveryTimeout: config.DeliveryTimeout,
	}, nil
}
