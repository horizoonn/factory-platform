package env

import (
	"errors"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"

	producerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/producer/franz"
)

type orderPaidProducerConfig struct {
	ClientID        string        `envconfig:"CLIENT_ID" default:"order-order-paid-producer"`
	Topic           string        `envconfig:"TOPIC" required:"true"`
	DeliveryTimeout time.Duration `envconfig:"DELIVERY_TIMEOUT" default:"10s"`
}

func NewOrderPaidProducerConfig(brokers []string) (producerfranz.Config, string, error) {
	var config orderPaidProducerConfig

	if err := envconfig.Process("ORDER_PAID_PRODUCER", &config); err != nil {
		return producerfranz.Config{}, "", fmt.Errorf("process order paid producer envconfig: %w", err)
	}
	if config.DeliveryTimeout <= 0 {
		return producerfranz.Config{}, "", errors.New("order paid producer delivery timeout must be positive")
	}

	return producerfranz.Config{
		Brokers:         brokers,
		ClientID:        config.ClientID,
		DeliveryTimeout: config.DeliveryTimeout,
	}, config.Topic, nil
}
