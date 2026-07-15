package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/horizoonn/factory-platform/assembly/internal/outbox"
	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

func (r *Repository) ClaimOne(
	ctx context.Context,
	workerID string,
	leaseDuration time.Duration,
) (outbox.Event, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		WITH candidate AS (
			SELECT id
			FROM platform.assembly_outbox_events
			WHERE published_at IS NULL
				AND failed_at IS NULL
				AND available_at <= NOW()
				AND next_attempt_at <= NOW()
				AND (locked_until IS NULL OR locked_until < NOW())
			ORDER BY available_at, created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE platform.assembly_outbox_events AS event
		SET locked_by = $1,
			locked_until = NOW() + ($2 * INTERVAL '1 second'),
			attempts = attempts + 1
		FROM candidate
		WHERE event.id = candidate.id
		RETURNING event.id, event.source_event_id, event.aggregate_id,
			event.topic, event.message_key, event.payload, event.headers,
			event.available_at, event.attempts
	`

	var (
		event       outbox.Event
		headersJSON []byte
	)

	err := r.pool.QueryRow(ctx, query, workerID, leaseDuration.Seconds()).Scan(
		&event.ID,
		&event.SourceEventID,
		&event.AggregateID,
		&event.Topic,
		&event.Key,
		&event.Payload,
		&headersJSON,
		&event.AvailableAt,
		&event.Attempts,
	)
	if errors.Is(err, postgrespool.ErrNoRows) {
		return outbox.Event{}, false, nil
	}
	if err != nil {
		return outbox.Event{}, false, fmt.Errorf("claim outbox event: %w", err)
	}

	if err := json.Unmarshal(headersJSON, &event.Headers); err != nil {
		return outbox.Event{}, false, fmt.Errorf("unmarshal outbox event headers: %w", err)
	}

	return event, true, nil
}
