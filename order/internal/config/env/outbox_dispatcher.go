package env

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"

	outboxdispatcher "github.com/horizoonn/factory-platform/order/internal/outbox/dispatcher"
)

type outboxConfig struct {
	PollInterval   time.Duration `envconfig:"POLL_INTERVAL" default:"250ms"`
	LeaseDuration  time.Duration `envconfig:"LEASE_DURATION" default:"30s"`
	PublishTimeout time.Duration `envconfig:"PUBLISH_TIMEOUT" default:"10s"`
	BaseBackoff    time.Duration `envconfig:"BASE_BACKOFF" default:"1s"`
	MaxBackoff     time.Duration `envconfig:"MAX_BACKOFF" default:"1m"`
	MaxAttempts    int           `envconfig:"MAX_ATTEMPTS" default:"10"`
}

func NewOutboxDispatcherConfig() (outboxdispatcher.Config, error) {
	var config outboxConfig

	if err := envconfig.Process("OUTBOX", &config); err != nil {
		return outboxdispatcher.Config{}, fmt.Errorf("process outbox envconfig: %w", err)
	}

	return outboxdispatcher.Config{
		PollInterval:   config.PollInterval,
		LeaseDuration:  config.LeaseDuration,
		PublishTimeout: config.PublishTimeout,
		BaseBackoff:    config.BaseBackoff,
		MaxBackoff:     config.MaxBackoff,
		MaxAttempts:    config.MaxAttempts,
	}, nil
}
