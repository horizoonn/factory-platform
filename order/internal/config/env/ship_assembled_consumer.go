package env

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	consumerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer/franz"
)

type shipAssembledConsumerConfig struct {
	ClientID       string `envconfig:"CLIENT_ID" default:"order-ship-assembled-consumer"`
	GroupID        string `envconfig:"GROUP_ID" default:"order-service"`
	Topic          string `envconfig:"TOPIC" required:"true"`
	MaxPollRecords int    `envconfig:"MAX_POLL_RECORDS" default:"100"`
}

func NewShipAssembledConsumerConfig(brokers []string) (consumerfranz.Config, error) {
	var config shipAssembledConsumerConfig

	if err := envconfig.Process("SHIP_ASSEMBLED_CONSUMER", &config); err != nil {
		return consumerfranz.Config{}, fmt.Errorf("process ship assembled consumer envconfig: %w", err)
	}

	return consumerfranz.Config{
		Brokers:        brokers,
		ClientID:       config.ClientID,
		GroupID:        config.GroupID,
		Topics:         []string{config.Topic},
		MaxPollRecords: config.MaxPollRecords,
	}, nil
}
