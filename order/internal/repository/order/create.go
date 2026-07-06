package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/repository/converter"
	"github.com/horizoonn/factory-platform/order/internal/repository/model"
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

	orderModel := converter.DomainOrderToModel(order)

	row := r.pool.QueryRow(
		ctx,
		query,
		orderModel.ID,
		orderModel.UserID,
		orderModel.PartIDs,
		orderModel.TotalPrice,
		orderModel.TransactionID,
		orderModel.PaymentMethod,
		orderModel.Status,
	)

	var result model.Order
	if err := result.Scan(row); err != nil {
		if errors.Is(err, postgrespool.ErrViolatesForeignKey) {
			return domain.Order{}, fmt.Errorf("referenced entity not found: %w", domain.ErrPartsNotFound)
		}
		return domain.Order{}, fmt.Errorf("scan created order: %w", err)
	}

	return converter.OrderModelToDomain(result)
}
