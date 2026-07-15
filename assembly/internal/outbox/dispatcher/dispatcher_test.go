package dispatcher

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/assembly/internal/outbox"
	"github.com/horizoonn/factory-platform/assembly/internal/outbox/mocks"
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
	errPublish := errors.New("non-retryable publish failure")
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

const dispatcherWorkerID = "assembly-worker-1"

var dispatcherNow = time.Date(2026, time.July, 11, 10, 0, 0, 0, time.UTC)

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
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f987001"),
		SourceEventID: uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f987002"),
		AggregateID:   uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f987003"),
		Topic:         "assembly.ship-assembled.v1",
		Key:           []byte("7d4a1f4f-07cc-48b2-b7c7-f6201f987003"),
		Payload:       []byte("payload"),
		Headers:       map[string]string{"event-type": "events.v1.ShipAssembled"},
		AvailableAt:   dispatcherNow,
		Attempts:      attempts,
	}
}
