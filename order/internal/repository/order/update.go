package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

func (r *Repository) UpdateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE platform.orders
	SET status=$1,
		transaction_id=$2,
		payment_method=$3
	WHERE id=$4
	RETURNING id, user_id, part_ids, total_price, transaction_id, payment_method, status, created_at, updated_at
	`

	model := domainOrderToModel(order)

	row := r.pool.QueryRow(
		ctx,
		query,
		model.Status,
		model.TransactionID,
		model.PaymentMethod,
		model.ID,
	)

	var result orderModel
	if err := result.scan(row); err != nil {
		if errors.Is(err, postgrespool.ErrNoRows) {
			return domain.Order{}, fmt.Errorf("order with id=%s: %w", order.ID, domain.ErrNotFound)
		}

		return domain.Order{}, fmt.Errorf("scan updated order: %w", err)
	}

	updatedOrder, err := result.toDomain()
	if err != nil {
		return domain.Order{}, fmt.Errorf("convert order model to domain: %w", err)
	}

	return updatedOrder, nil
}
