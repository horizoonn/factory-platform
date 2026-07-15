package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (r *Repository) MarkPublished(
	ctx context.Context,
	workerID string,
	eventID uuid.UUID,
) error {
	const query = `
		UPDATE platform.order_outbox_events
		SET published_at = NOW(),
			locked_by = NULL,
			locked_until = NULL,
			last_error = NULL
		WHERE id = $1
			AND locked_by = $2
			AND published_at IS NULL
			AND failed_at IS NULL
	`

	return r.execLeaseUpdate(ctx, query, eventID, workerID)
}

func (r *Repository) Reschedule(
	ctx context.Context,
	workerID string,
	eventID uuid.UUID,
	nextAttemptAt time.Time,
	lastError string,
) error {
	const query = `
		UPDATE platform.order_outbox_events
		SET next_attempt_at = $3,
			last_error = $4,
			locked_by = NULL,
			locked_until = NULL
		WHERE id = $1
			AND locked_by = $2
			AND published_at IS NULL
			AND failed_at IS NULL
	`

	return r.execLeaseUpdate(
		ctx,
		query,
		eventID,
		workerID,
		nextAttemptAt,
		lastError,
	)
}

func (r *Repository) MarkFailed(
	ctx context.Context,
	workerID string,
	eventID uuid.UUID,
	lastError string,
) error {
	const query = `
		UPDATE platform.order_outbox_events
		SET failed_at = NOW(),
			last_error = $3,
			locked_by = NULL,
			locked_until = NULL
		WHERE id = $1
			AND locked_by = $2
			AND published_at IS NULL
			AND failed_at IS NULL
	`

	return r.execLeaseUpdate(
		ctx,
		query,
		eventID,
		workerID,
		lastError,
	)
}

func (r *Repository) execLeaseUpdate(
	ctx context.Context,
	query string,
	eventID uuid.UUID,
	workerID string,
	args ...any,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	queryArgs := make([]any, 0, len(args)+2)
	queryArgs = append(queryArgs, eventID, workerID)
	queryArgs = append(queryArgs, args...)

	tag, err := r.pool.Exec(ctx, query, queryArgs...)
	if err != nil {
		return fmt.Errorf("update outbox event lease: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ErrLeaseLost
	}

	return nil
}
