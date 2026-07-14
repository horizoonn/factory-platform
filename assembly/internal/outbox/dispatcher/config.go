package dispatcher

import (
	"errors"
	"strings"
	"time"
)

type Config struct {
	WorkerID       string
	PollInterval   time.Duration
	LeaseDuration  time.Duration
	PublishTimeout time.Duration
	BaseBackoff    time.Duration
	MaxBackoff     time.Duration
	MaxAttempts    int
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.WorkerID) == "" {
		return errors.New("outbox worker id is required")
	}
	if c.PollInterval <= 0 {
		return errors.New("outbox poll interval must be positive")
	}
	if c.PublishTimeout <= 0 {
		return errors.New("outbox publish timeout must be positive")
	}
	if c.LeaseDuration <= c.PublishTimeout {
		return errors.New("outbox lease duration must exceed publish timeout")
	}
	if c.BaseBackoff <= 0 {
		return errors.New("outbox base backoff must be positive")
	}
	if c.MaxBackoff < c.BaseBackoff {
		return errors.New("outbox max backoff must not be less than base backoff")
	}
	if c.MaxAttempts <= 0 {
		return errors.New("outbox max attempts must be positive")
	}

	return nil
}
