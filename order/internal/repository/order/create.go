package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

func (r *Repository) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO platform.orders (id, user_id, part_ids, total_price, transaction_id, payment_method, status)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, user_id, part_ids, total_price, transaction_id, payment_method, status, created_at, updated_at;
	`

	model := domainOrderToModel(order)

	row := r.pool.QueryRow(
		ctx,
		query,
		model.ID,
		model.UserID,
		model.PartIDs,
		model.TotalPrice,
		model.TransactionID,
		model.PaymentMethod,
		model.Status,
	)

	var result orderModel
	if err := result.scan(row); err != nil {
		if errors.Is(err, postgrespool.ErrViolatesForeignKey) {
			return domain.Order{}, fmt.Errorf("referenced entity not found: %w", domain.ErrPartsNotFound)
		}
		return domain.Order{}, fmt.Errorf("scan created order: %w", err)
	}

	return result.toDomain()
}
