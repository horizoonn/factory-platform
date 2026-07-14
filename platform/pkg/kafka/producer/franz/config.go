package franz

import (
	"errors"
	"strings"
	"time"
)

const defaultDeliveryTimeout = 10 * time.Second

type Config struct {
	Brokers         []string
	ClientID        string
	DeliveryTimeout time.Duration
}

func (c Config) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("kafka brokers are required")
	}

	for _, broker := range c.Brokers {
		if strings.TrimSpace(broker) == "" {
			return errors.New("kafka broker address is required")
		}
	}
	if c.DeliveryTimeout < 0 {
		return errors.New("kafka delivery timeout must not be negative")
	}

	return nil
}

func (c Config) deliveryTimeout() time.Duration {
	if c.DeliveryTimeout <= 0 {
		return defaultDeliveryTimeout
	}

	return c.DeliveryTimeout
}
