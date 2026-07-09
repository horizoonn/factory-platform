package franz

import (
	"errors"
	"strings"
)

type Config struct {
	Brokers      []string
	ClientID     string
	DefaultTopic string
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

	return nil
}
