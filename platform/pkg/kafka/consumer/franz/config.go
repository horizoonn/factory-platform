package franz

import (
	"errors"
	"strings"
)

const defaultMaxPollRecords = 100

type Config struct {
	Brokers        []string
	ClientID       string
	GroupID        string
	Topics         []string
	MaxPollRecords int
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

	if strings.TrimSpace(c.GroupID) == "" {
		return errors.New("kafka consumer group id is required")
	}

	if len(c.Topics) == 0 {
		return errors.New("kafka topics are required")
	}

	for _, topic := range c.Topics {
		if strings.TrimSpace(topic) == "" {
			return errors.New("kafka topic is required")
		}
	}

	return nil
}

func (c Config) pollLimit() int {
	if c.MaxPollRecords <= 0 {
		return defaultMaxPollRecords
	}

	return c.MaxPollRecords
}
