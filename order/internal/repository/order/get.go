package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

func (r *Repository) GetOrder(ctx context.Context, id uuid.UUID) (domain.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, user_id, part_ids, total_price, transaction_id, payment_method, status, created_at, updated_at
	FROM platform.orders
	WHERE id=$1;
	`

	row := r.pool.QueryRow(ctx, query, id)

	var model orderModel
	if err := model.scan(row); err != nil {
		if errors.Is(err, postgrespool.ErrNoRows) {
			return domain.Order{}, fmt.Errorf("order with id=%s: %w", id, domain.ErrNotFound)
		}

		return domain.Order{}, fmt.Errorf("scan order row: %w", err)
	}

	order, err := model.toDomain()
	if err != nil {
		return domain.Order{}, fmt.Errorf("convert order model to domain: %w", err)
	}

	return order, nil
}
