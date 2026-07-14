//go:build integration

package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	outboxmodel "github.com/horizoonn/factory-platform/assembly/internal/outbox"
	outboxrepository "github.com/horizoonn/factory-platform/assembly/internal/repository/outbox"
)

func TestOutboxRepositoryEnqueue(t *testing.T) {
	testEnv.truncateOutbox(t)

	repository := outboxrepository.NewRepository(testEnv.pool)
	event := outboxFixture()

	created, err := repository.Enqueue(testContext(t), event)

	require.NoError(t, err)
	assert.True(t, created)

	var (
		id            uuid.UUID
		sourceEventID uuid.UUID
		aggregateID   uuid.UUID
		topic         string
		key           []byte
		payload       []byte
		headersJSON   []byte
		availableAt   time.Time
		nextAttemptAt time.Time
	)

	err = testEnv.pool.QueryRow(testContext(t), `
		SELECT id, source_event_id, aggregate_id, topic, message_key,
			payload, headers, available_at, next_attempt_at
		FROM platform.assembly_outbox_events
		WHERE id = $1
	`, event.ID).Scan(
		&id,
		&sourceEventID,
		&aggregateID,
		&topic,
		&key,
		&payload,
		&headersJSON,
		&availableAt,
		&nextAttemptAt,
	)
	require.NoError(t, err)

	var headers map[string]string
	require.NoError(t, json.Unmarshal(headersJSON, &headers))

	assert.Equal(t, event.ID, id)
	assert.Equal(t, event.SourceEventID, sourceEventID)
	assert.Equal(t, event.AggregateID, aggregateID)
	assert.Equal(t, event.Topic, topic)
	assert.Equal(t, event.Key, key)
	assert.Equal(t, event.Payload, payload)
	assert.Equal(t, event.Headers, headers)
	assert.True(t, event.AvailableAt.Equal(availableAt))
	assert.True(t, event.AvailableAt.Equal(nextAttemptAt))
}

func TestOutboxRepositoryEnqueueDuplicateSourceEvent(t *testing.T) {
	testEnv.truncateOutbox(t)

	repository := outboxrepository.NewRepository(testEnv.pool)
	first := outboxFixture()
	duplicate := first
	duplicate.ID = uuid.New()
	duplicate.Payload = []byte("different payload")

	created, err := repository.Enqueue(testContext(t), first)
	require.NoError(t, err)
	assert.True(t, created)

	created, err = repository.Enqueue(testContext(t), duplicate)
	require.NoError(t, err)
	assert.False(t, created)

	var count int
	err = testEnv.pool.QueryRow(
		testContext(t),
		"SELECT COUNT(*) FROM platform.assembly_outbox_events WHERE source_event_id = $1",
		first.SourceEventID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestOutboxRepositoryClaimAndMarkPublished(t *testing.T) {
	testEnv.truncateOutbox(t)

	repository := outboxrepository.NewRepository(testEnv.pool)
	event := outboxFixture()
	event.AvailableAt = time.Now().UTC().Add(-time.Second)

	created, err := repository.Enqueue(testContext(t), event)
	require.NoError(t, err)
	require.True(t, created)

	claimed, found, err := repository.ClaimOne(testContext(t), "worker-1", 30*time.Second)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, event.ID, claimed.ID)
	assert.Equal(t, 1, claimed.Attempts)

	_, found, err = repository.ClaimOne(testContext(t), "worker-2", 30*time.Second)
	require.NoError(t, err)
	assert.False(t, found)

	err = repository.MarkPublished(testContext(t), "worker-2", event.ID)
	require.ErrorIs(t, err, outboxrepository.ErrLeaseLost)

	require.NoError(t, repository.MarkPublished(testContext(t), "worker-1", event.ID))

	_, found, err = repository.ClaimOne(testContext(t), "worker-2", 30*time.Second)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestOutboxRepositoryRescheduleAndMarkFailed(t *testing.T) {
	testEnv.truncateOutbox(t)

	repository := outboxrepository.NewRepository(testEnv.pool)
	event := outboxFixture()
	event.AvailableAt = time.Now().UTC().Add(-time.Second)

	created, err := repository.Enqueue(testContext(t), event)
	require.NoError(t, err)
	require.True(t, created)

	claimed, found, err := repository.ClaimOne(testContext(t), "worker-1", 30*time.Second)
	require.NoError(t, err)
	require.True(t, found)

	nextAttemptAt := time.Now().UTC().Add(time.Hour)
	require.NoError(t, repository.Reschedule(
		testContext(t),
		"worker-1",
		claimed.ID,
		nextAttemptAt,
		"kafka unavailable",
	))

	_, found, err = repository.ClaimOne(testContext(t), "worker-2", 30*time.Second)
	require.NoError(t, err)
	assert.False(t, found)

	_, err = testEnv.pool.Exec(
		testContext(t),
		"UPDATE platform.assembly_outbox_events SET next_attempt_at = NOW() - INTERVAL '1 second' WHERE id = $1",
		event.ID,
	)
	require.NoError(t, err)

	claimed, found, err = repository.ClaimOne(testContext(t), "worker-2", 30*time.Second)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, 2, claimed.Attempts)

	require.NoError(t, repository.MarkFailed(
		testContext(t),
		"worker-2",
		claimed.ID,
		"attempt limit reached",
	))

	_, found, err = repository.ClaimOne(testContext(t), "worker-3", 30*time.Second)
	require.NoError(t, err)
	assert.False(t, found)

	var failedAt *time.Time
	err = testEnv.pool.QueryRow(
		testContext(t),
		"SELECT failed_at FROM platform.assembly_outbox_events WHERE id = $1",
		event.ID,
	).Scan(&failedAt)
	require.NoError(t, err)
	require.NotNil(t, failedAt)
}

func TestOutboxRepositoryExpiredLeaseCanBeReclaimed(t *testing.T) {
	testEnv.truncateOutbox(t)

	repository := outboxrepository.NewRepository(testEnv.pool)
	event := outboxFixture()
	event.AvailableAt = time.Now().UTC().Add(-time.Second)

	created, err := repository.Enqueue(testContext(t), event)
	require.NoError(t, err)
	require.True(t, created)

	claimed, found, err := repository.ClaimOne(testContext(t), "worker-1", 30*time.Second)
	require.NoError(t, err)
	require.True(t, found)

	_, err = testEnv.pool.Exec(
		testContext(t),
		"UPDATE platform.assembly_outbox_events SET locked_until = NOW() - INTERVAL '1 second' WHERE id = $1",
		event.ID,
	)
	require.NoError(t, err)

	reclaimed, found, err := repository.ClaimOne(testContext(t), "worker-2", 30*time.Second)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, claimed.ID, reclaimed.ID)
	assert.Equal(t, 2, reclaimed.Attempts)

	err = repository.MarkPublished(testContext(t), "worker-1", event.ID)
	require.ErrorIs(t, err, outboxrepository.ErrLeaseLost)
	require.NoError(t, repository.MarkPublished(testContext(t), "worker-2", event.ID))
}

func outboxFixture() outboxmodel.Event {
	availableAt := time.Now().UTC().Add(10 * time.Second).Truncate(time.Microsecond)

	return outboxmodel.Event{
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f983001"),
		SourceEventID: uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f983002"),
		AggregateID:   uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f983003"),
		Topic:         "assembly.ship-assembled.v1",
		Key:           []byte("7d4a1f4f-07cc-48b2-b7c7-f6201f983003"),
		Payload:       []byte("ship assembled payload"),
		Headers:       map[string]string{"event-type": "events.v1.ShipAssembled"},
		AvailableAt:   availableAt,
	}
}
