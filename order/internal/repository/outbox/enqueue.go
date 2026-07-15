package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/outbox"
	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

func (r *Repository) Enqueue(
	ctx context.Context,
	executor postgrespool.Executor,
	event outbox.Event,
) (bool, error) {
	headers := event.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	headersJSON, err := json.Marshal(headers)
	if err != nil {
		return false, fmt.Errorf("marshal outbox event headers: %w", err)
	}

	const query = `
		INSERT INTO platform.order_outbox_events (
			id,
			aggregate_id,
			event_type,
			topic,
			message_key,
			payload,
			headers,
			available_at,
			next_attempt_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $8
		)
		ON CONFLICT (aggregate_id, event_type)
		DO NOTHING
	`

	tag, err := executor.Exec(
		ctx,
		query,
		event.ID,
		event.AggregateID,
		event.Type,
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
