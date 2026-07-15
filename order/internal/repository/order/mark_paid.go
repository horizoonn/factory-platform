package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/outbox"
	"github.com/horizoonn/factory-platform/order/internal/repository/converter"
	"github.com/horizoonn/factory-platform/order/internal/repository/model"
	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

func (r *Repository) MarkPaidAndEnqueueOrderPaid(
	ctx context.Context,
	order domain.Order,
	event outbox.Event,
) (domain.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var updatedOrder domain.Order

	err := r.pool.WithinTransaction(ctx, func(tx postgrespool.Executor) error {
		var err error
		updatedOrder, err = markOrderPaid(ctx, tx, order)
		if err != nil {
			return fmt.Errorf("update paid order: %w", err)
		}

		created, err := r.outbox.Enqueue(ctx, tx, event)
		if err != nil {
			return fmt.Errorf("enqueue OrderPaid event: %w", err)
		}

		if !created {
			return fmt.Errorf(
				"OrderPaid event for order id=%s already exists: %w",
				order.ID,
				domain.ErrOrderAlreadyPaid,
			)
		}

		return nil
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf(
			"mark order paid and enqueue event: %w",
			err,
		)
	}

	return updatedOrder, nil
}

func markOrderPaid(
	ctx context.Context,
	executor postgrespool.Executor,
	order domain.Order,
) (domain.Order, error) {
	const query = `
		UPDATE platform.orders
		SET
			status = $1,
			transaction_id = $2,
			payment_method = $3
		WHERE id = $4
			AND status = $5
		RETURNING
			id,
			user_id,
			part_ids,
			total_price,
			transaction_id,
			payment_method,
			status,
			created_at,
			updated_at
	`

	orderModel := converter.DomainOrderToModel(order)
	row := executor.QueryRow(
		ctx,
		query,
		orderModel.Status,
		orderModel.TransactionID,
		orderModel.PaymentMethod,
		orderModel.ID,
		domain.OrderStatusPendingPayment,
	)

	var result model.Order
	if err := result.Scan(row); err != nil {
		if errors.Is(err, postgrespool.ErrNoRows) {
			status, statusErr := orderStatus(ctx, executor, order.ID)
			if statusErr != nil {
				return domain.Order{}, statusErr
			}

			switch status {
			case domain.OrderStatusPaid, domain.OrderStatusCompleted:
				return domain.Order{}, fmt.Errorf(
					"order id=%s has status=%s: %w",
					order.ID,
					status,
					domain.ErrOrderAlreadyPaid,
				)
			case domain.OrderStatusCancelled:
				return domain.Order{}, fmt.Errorf("order id=%s: %w", order.ID, domain.ErrOrderCancelled)
			default:
				return domain.Order{}, fmt.Errorf(
					"order id=%s has status=%s: %w",
					order.ID,
					status,
					domain.ErrInvalidOrderStatus,
				)
			}
		}

		return domain.Order{}, fmt.Errorf("scan paid order id=%s: %w", order.ID, err)
	}

	updatedOrder, err := converter.OrderModelToDomain(result)
	if err != nil {
		return domain.Order{}, fmt.Errorf("convert paid order id=%s to domain: %w", order.ID, err)
	}

	return updatedOrder, nil
}
