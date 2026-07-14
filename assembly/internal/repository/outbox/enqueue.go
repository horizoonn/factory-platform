package repository

import (
	"context"
	"encoding/json"
	"fmt"

	outboxmodel "github.com/horizoonn/factory-platform/assembly/internal/outbox"
)

func (r *Repository) Enqueue(ctx context.Context, event outboxmodel.Event) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	headers := event.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	headersJSON, err := json.Marshal(headers)
	if err != nil {
		return false, fmt.Errorf("marshal outbox event headers: %w", err)
	}

	const query = `
		INSERT INTO platform.assembly_outbox_events (
			id, source_event_id, aggregate_id, topic, message_key,
			payload, headers, available_at, next_attempt_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		ON CONFLICT (source_event_id) DO NOTHING
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		event.ID,
		event.SourceEventID,
		event.AggregateID,
		event.Topic,
		event.Key,
		event.Payload,
		headersJSON,
		event.AvailableAt,
	)
	if err != nil {
		return false, fmt.Errorf("enqueue outbox event: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}
