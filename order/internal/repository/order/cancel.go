package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

func (r *Repository) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		UPDATE platform.orders
		SET status = $1
		WHERE id = $2
			AND status = $3
	`

	tag, err := r.pool.Exec(
		ctx,
		query,
		domain.OrderStatusCancelled,
		orderID,
		domain.OrderStatusPendingPayment,
	)
	if err != nil {
		return fmt.Errorf("cancel order id=%s: %w", orderID, err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	status, err := orderStatus(ctx, r.pool, orderID)
	if err != nil {
		return err
	}

	switch status {
	case domain.OrderStatusCancelled:
		return nil
	case domain.OrderStatusPaid, domain.OrderStatusCompleted:
		return fmt.Errorf("order id=%s has status=%s: %w", orderID, status, domain.ErrOrderAlreadyPaid)
	default:
		return fmt.Errorf(
			"order id=%s has status=%s: %w",
			orderID,
			status,
			domain.ErrInvalidOrderStatus,
		)
	}
}

func orderStatus(
	ctx context.Context,
	executor postgrespool.Executor,
	orderID uuid.UUID,
) (domain.OrderStatus, error) {
	const query = `
		SELECT status
		FROM platform.orders
		WHERE id = $1
	`

	var status domain.OrderStatus
	if err := executor.QueryRow(ctx, query, orderID).Scan(&status); err != nil {
		if errors.Is(err, postgrespool.ErrNoRows) {
			return domain.OrderStatusUnknown, fmt.Errorf("order id=%s: %w", orderID, domain.ErrNotFound)
		}

		return domain.OrderStatusUnknown, fmt.Errorf("scan order status id=%s: %w", orderID, err)
	}

	return status, nil
}
