package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

const shipAssembledEventType = "events.v1.ShipAssembled" //nolint:gosec // G101: Kafka event type, not a credential.

func (r *Repository) CompleteOrder(ctx context.Context, event domain.ShipAssembledEvent) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	err := r.pool.WithinTransaction(ctx, func(tx postgrespool.Executor) error {
		created, err := registerShipAssembledEvent(ctx, tx, event)
		if err != nil {
			return err
		}
		if !created {
			return nil
		}

		completed, err := markOrderCompleted(ctx, tx, event)
		if err != nil {
			return err
		}
		if completed {
			return nil
		}

		return completeOrderConflict(ctx, tx, event)
	})
	if err != nil {
		return fmt.Errorf("complete order transaction: %w", err)
	}

	return nil
}

func registerShipAssembledEvent(
	ctx context.Context,
	executor postgrespool.Executor,
	event domain.ShipAssembledEvent,
) (bool, error) {
	const query = `
		INSERT INTO platform.order_inbox_events (
			event_id,
			aggregate_id,
			event_type
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (event_id) DO NOTHING
	`

	tag, err := executor.Exec(
		ctx,
		query,
		event.ID,
		event.OrderID,
		shipAssembledEventType,
	)
	if err != nil {
		return false, fmt.Errorf("register ShipAssembled inbox event: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func markOrderCompleted(
	ctx context.Context,
	executor postgrespool.Executor,
	event domain.ShipAssembledEvent,
) (bool, error) {
	const query = `
		UPDATE platform.orders
		SET status = $1
		WHERE id = $2
			AND user_id = $3
			AND status = $4
	`

	tag, err := executor.Exec(
		ctx,
		query,
		domain.OrderStatusCompleted,
		event.OrderID,
		event.UserID,
		domain.OrderStatusPaid,
	)
	if err != nil {
		return false, fmt.Errorf("mark order completed: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func completeOrderConflict(
	ctx context.Context,
	executor postgrespool.Executor,
	event domain.ShipAssembledEvent,
) error {
	const query = `
		SELECT user_id, status
		FROM platform.orders
		WHERE id = $1
	`

	var (
		userID uuid.UUID
		status domain.OrderStatus
	)
	if err := executor.QueryRow(ctx, query, event.OrderID).Scan(&userID, &status); err != nil {
		if errors.Is(err, postgrespool.ErrNoRows) {
			return fmt.Errorf("order id=%s: %w", event.OrderID, domain.ErrNotFound)
		}

		return fmt.Errorf("scan order completion conflict: %w", err)
	}

	if userID != event.UserID {
		return fmt.Errorf(
			"ShipAssembled user=%s does not own order=%s: %w",
			event.UserID,
			event.OrderID,
			domain.ErrOrderUserMismatch,
		)
	}
	if status == domain.OrderStatusCompleted {
		return nil
	}

	return fmt.Errorf(
		"cannot complete order=%s in status=%s: %w",
		event.OrderID,
		status,
		domain.ErrInvalidOrderStatus,
	)
}
