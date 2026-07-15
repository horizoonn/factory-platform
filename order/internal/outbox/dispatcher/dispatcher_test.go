package dispatcher

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/order/internal/outbox"
	"github.com/horizoonn/factory-platform/order/internal/outbox/mocks"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
)

func TestDispatcher_ProcessOne_Published(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	event := dispatcherEventFixture(1)
	repository := mocks.NewRepository(t)
	producer := mocks.NewProducer(t)
	dispatcher := newTestDispatcher(t, repository, producer)

	repository.EXPECT().
		ClaimOne(ctx, dispatcherWorkerID, dispatcherConfig().LeaseDuration).
		Return(event, true, nil).
		Once()
	producer.EXPECT().
		Publish(mock.Anything, event.Topic, kafka.Message{
			Key:     event.Key,
			Value:   event.Payload,
			Headers: kafka.TextHeaders(event.Headers),
		}).
		Return(nil).
		Once()
	repository.EXPECT().
		MarkPublished(ctx, dispatcherWorkerID, event.ID).
		Return(nil).
		Once()

	processed, err := dispatcher.processOne(ctx)

	require.NoError(t, err)
	assert.True(t, processed)
}

func TestDispatcher_ProcessOne_NoEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repository := mocks.NewRepository(t)
	producer := mocks.NewProducer(t)
	dispatcher := newTestDispatcher(t, repository, producer)

	repository.EXPECT().
		ClaimOne(ctx, dispatcherWorkerID, dispatcherConfig().LeaseDuration).
		Return(outbox.Event{}, false, nil).
		Once()

	processed, err := dispatcher.processOne(ctx)

	require.NoError(t, err)
	assert.False(t, processed)
}

func TestDispatcher_ProcessOne_ClaimError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	errRepository := errors.New("postgres unavailable")
	repository := mocks.NewRepository(t)
	producer := mocks.NewProducer(t)
	dispatcher := newTestDispatcher(t, repository, producer)

	repository.EXPECT().
		ClaimOne(ctx, dispatcherWorkerID, dispatcherConfig().LeaseDuration).
		Return(outbox.Event{}, false, errRepository).
		Once()

	processed, err := dispatcher.processOne(ctx)

	require.ErrorIs(t, err, errRepository)
	assert.False(t, processed)
}

func TestDispatcher_ProcessOne_Rescheduled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	event := dispatcherEventFixture(2)
	errPublish := errors.New("kafka unavailable")
	repository := mocks.NewRepository(t)
	producer := mocks.NewProducer(t)
	dispatcher := newTestDispatcher(t, repository, producer)
	nextAttemptAt := dispatcherNow.Add(2 * time.Second)

	repository.EXPECT().
		ClaimOne(ctx, dispatcherWorkerID, dispatcherConfig().LeaseDuration).
		Return(event, true, nil).
		Once()
	producer.EXPECT().
		Publish(mock.Anything, event.Topic, kafka.Message{
			Key:     event.Key,
			Value:   event.Payload,
			Headers: kafka.TextHeaders(event.Headers),
		}).
		Return(errPublish).
		Once()
	repository.EXPECT().
		Reschedule(ctx, dispatcherWorkerID, event.ID, nextAttemptAt, errPublish.Error()).
		Return(nil).
		Once()

	processed, err := dispatcher.processOne(ctx)

	require.NoError(t, err)
	assert.True(t, processed)
}

func TestDispatcher_ProcessOne_MaxAttempts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := dispatcherConfig()
	event := dispatcherEventFixture(config.MaxAttempts)
	errPublish := errors.New("publish failed")
	repository := mocks.NewRepository(t)
	producer := mocks.NewProducer(t)
	dispatcher := newTestDispatcher(t, repository, producer)

	repository.EXPECT().
		ClaimOne(ctx, dispatcherWorkerID, config.LeaseDuration).
		Return(event, true, nil).
		Once()
	producer.EXPECT().
		Publish(mock.Anything, event.Topic, kafka.Message{
			Key:     event.Key,
			Value:   event.Payload,
			Headers: kafka.TextHeaders(event.Headers),
		}).
		Return(errPublish).
		Once()
	repository.EXPECT().
		MarkFailed(ctx, dispatcherWorkerID, event.ID, errPublish.Error()).
		Return(nil).
		Once()

	processed, err := dispatcher.processOne(ctx)

	require.NoError(t, err)
	assert.True(t, processed)
}

func TestDispatcher_ProcessOne_MarkPublishedError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	event := dispatcherEventFixture(1)
	errRepository := errors.New("lease lost")
	repository := mocks.NewRepository(t)
	producer := mocks.NewProducer(t)
	dispatcher := newTestDispatcher(t, repository, producer)

	repository.EXPECT().
		ClaimOne(ctx, dispatcherWorkerID, dispatcherConfig().LeaseDuration).
		Return(event, true, nil).
		Once()
	producer.EXPECT().
		Publish(mock.Anything, event.Topic, mock.Anything).
		Return(nil).
		Once()
	repository.EXPECT().
		MarkPublished(ctx, dispatcherWorkerID, event.ID).
		Return(errRepository).
		Once()

	processed, err := dispatcher.processOne(ctx)

	require.ErrorIs(t, err, errRepository)
	assert.True(t, processed)
}

func TestDispatcher_Run_CanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repository := mocks.NewRepository(t)
	producer := mocks.NewProducer(t)
	dispatcher := newTestDispatcher(t, repository, producer)

	require.ErrorIs(t, dispatcher.Run(ctx), context.Canceled)
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "empty worker id", mutate: func(c *Config) { c.WorkerID = " " }},
		{name: "invalid poll interval", mutate: func(c *Config) { c.PollInterval = 0 }},
		{name: "invalid publish timeout", mutate: func(c *Config) { c.PublishTimeout = 0 }},
		{name: "lease does not exceed publish timeout", mutate: func(c *Config) { c.LeaseDuration = c.PublishTimeout }},
		{name: "invalid base backoff", mutate: func(c *Config) { c.BaseBackoff = 0 }},
		{name: "max backoff below base", mutate: func(c *Config) { c.MaxBackoff = c.BaseBackoff - time.Nanosecond }},
		{name: "invalid max attempts", mutate: func(c *Config) { c.MaxAttempts = 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := dispatcherConfig()
			tt.mutate(&config)

			require.Error(t, config.Validate())
		})
	}

	require.NoError(t, dispatcherConfig().Validate())
}

func TestRetryBackoff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 1, want: time.Second},
		{attempt: 2, want: 2 * time.Second},
		{attempt: 4, want: 8 * time.Second},
		{attempt: 10, want: time.Minute},
	}

	for _, tt := range tests {
		got := retryBackoff(tt.attempt, time.Second, time.Minute)
		assert.Equal(t, tt.want, got)
	}
}

func TestTruncateErrorUsesRunes(t *testing.T) {
	t.Parallel()

	message := strings.Repeat("я", maxStoredErrorRunes+1)

	truncated := truncateError(message)

	assert.Len(t, []rune(truncated), maxStoredErrorRunes)
}

const dispatcherWorkerID = "order-worker-1"

var dispatcherNow = time.Date(2026, time.July, 14, 18, 0, 0, 0, time.UTC)

func dispatcherConfig() Config {
	return Config{
		WorkerID:       dispatcherWorkerID,
		PollInterval:   250 * time.Millisecond,
		LeaseDuration:  30 * time.Second,
		PublishTimeout: 10 * time.Second,
		BaseBackoff:    time.Second,
		MaxBackoff:     time.Minute,
		MaxAttempts:    10,
	}
}

func newTestDispatcher(t *testing.T, repository Repository, producer Producer) *Dispatcher {
	t.Helper()

	dispatcher, err := NewDispatcher(repository, producer, dispatcherConfig())
	require.NoError(t, err)
	dispatcher.clock = func() time.Time { return dispatcherNow }

	return dispatcher
}

func dispatcherEventFixture(attempts int) outbox.Event {
	return outbox.Event{
		ID:          uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f987001"),
		AggregateID: uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f987002"),
		Type:        "events.v1.OrderPaid",
		Topic:       "order.paid.v1",
		Key:         []byte("7d4a1f4f-07cc-48b2-b7c7-f6201f987002"),
		Payload:     []byte("payload"),
		Headers:     map[string]string{"event-type": "events.v1.OrderPaid"},
		AvailableAt: dispatcherNow,
		Attempts:    attempts,
	}
}
