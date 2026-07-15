package dispatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/horizoonn/factory-platform/order/internal/outbox"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

const maxStoredErrorRunes = 4096

type Repository interface {
	ClaimOne(
		ctx context.Context,
		workerID string,
		leaseDuration time.Duration,
	) (outbox.Event, bool, error)
	MarkPublished(ctx context.Context, workerID string, eventID uuid.UUID) error
	Reschedule(
		ctx context.Context,
		workerID string,
		eventID uuid.UUID,
		nextAttemptAt time.Time,
		lastError string,
	) error
	MarkFailed(ctx context.Context, workerID string, eventID uuid.UUID, lastError string) error
}

type Producer interface {
	Publish(ctx context.Context, topic string, msg kafka.Message) error
}

type Clock func() time.Time

type Dispatcher struct {
	repository Repository
	producer   Producer
	config     Config
	clock      Clock
}

func NewDispatcher(repository Repository, producer Producer, config Config) (*Dispatcher, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate outbox dispatcher config: %w", err)
	}

	return &Dispatcher{
		repository: repository,
		producer:   producer,
		config:     config,
		clock:      time.Now,
	}, nil
}

func (d *Dispatcher) Run(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		processed, err := d.processOne(ctx)
		if err != nil {
			return err
		}
		if processed {
			continue
		}

		if err := wait(ctx, d.config.PollInterval); err != nil {
			return err
		}
	}
}

func (d *Dispatcher) processOne(ctx context.Context) (bool, error) {
	event, found, err := d.repository.ClaimOne(ctx, d.config.WorkerID, d.config.LeaseDuration)
	if err != nil {
		return false, fmt.Errorf("claim outbox event: %w", err)
	}
	if !found {
		return false, nil
	}

	publishCtx, cancel := context.WithTimeout(ctx, d.config.PublishTimeout)
	err = d.producer.Publish(publishCtx, event.Topic, kafka.Message{
		Key:     event.Key,
		Value:   event.Payload,
		Headers: kafka.TextHeaders(event.Headers),
	})
	cancel()

	if ctxErr := ctx.Err(); ctxErr != nil {
		return true, ctxErr
	}
	if err != nil {
		return true, d.handlePublishError(ctx, event, err)
	}

	if err := d.repository.MarkPublished(ctx, d.config.WorkerID, event.ID); err != nil {
		return true, fmt.Errorf("mark outbox event published: %w", err)
	}
	logger.Debug(
		ctx,
		"outbox event published",
		zap.String("event_id", event.ID.String()),
		zap.String("topic", event.Topic),
		zap.Int("attempt", event.Attempts),
	)

	return true, nil
}

func (d *Dispatcher) handlePublishError(
	ctx context.Context,
	event outbox.Event,
	publishErr error,
) error {
	lastError := truncateError(publishErr.Error())
	if event.Attempts >= d.config.MaxAttempts {
		if err := d.repository.MarkFailed(
			ctx,
			d.config.WorkerID,
			event.ID,
			lastError,
		); err != nil {
			return fmt.Errorf("mark outbox event failed: %w", err)
		}
		logger.Error(
			ctx,
			"outbox event exhausted retries",
			zap.String("event_id", event.ID.String()),
			zap.String("topic", event.Topic),
			zap.Int("attempt", event.Attempts),
			zap.String("publish_error", lastError),
		)

		return nil
	}

	nextAttemptAt := d.clock().UTC().Add(retryBackoff(
		event.Attempts,
		d.config.BaseBackoff,
		d.config.MaxBackoff,
	))
	if err := d.repository.Reschedule(
		ctx,
		d.config.WorkerID,
		event.ID,
		nextAttemptAt,
		lastError,
	); err != nil {
		return fmt.Errorf("reschedule outbox event: %w", err)
	}
	logger.Warn(
		ctx,
		"outbox event publish failed; retry scheduled",
		zap.String("event_id", event.ID.String()),
		zap.String("topic", event.Topic),
		zap.Int("attempt", event.Attempts),
		zap.Time("next_attempt_at", nextAttemptAt),
		zap.String("publish_error", lastError),
	)

	return nil
}

func retryBackoff(attempt int, base, maximum time.Duration) time.Duration {
	backoff := base
	for current := 1; current < attempt && backoff < maximum; current++ {
		if backoff > maximum/2 {
			return maximum
		}
		backoff *= 2
	}
	if backoff > maximum {
		return maximum
	}

	return backoff
}

func truncateError(message string) string {
	runes := []rune(message)
	if len(runes) <= maxStoredErrorRunes {
		return message
	}

	return string(runes[:maxStoredErrorRunes])
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
