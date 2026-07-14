package env

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	consumerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer/franz"
)

type orderPaidConsumerConfig struct {
	ClientID       string `envconfig:"CLIENT_ID" default:"assembly-order-paid-consumer"`
	GroupID        string `envconfig:"GROUP_ID" default:"assembly-service"`
	Topic          string `envconfig:"TOPIC" required:"true"`
	MaxPollRecords int    `envconfig:"MAX_POLL_RECORDS" default:"100"`
}

func NewOrderPaidConsumerConfig(brokers []string) (consumerfranz.Config, error) {
	var config orderPaidConsumerConfig

	if err := envconfig.Process("ORDER_PAID_CONSUMER", &config); err != nil {
		return consumerfranz.Config{}, fmt.Errorf("process order paid consumer envconfig: %w", err)
	}

	return consumerfranz.Config{
		Brokers:        brokers,
		ClientID:       config.ClientID,
		GroupID:        config.GroupID,
		Topics:         []string{config.Topic},
		MaxPollRecords: config.MaxPollRecords,
	}, nil
}
