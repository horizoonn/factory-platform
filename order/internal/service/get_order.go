package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
)

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (domain.Order, error) {
	if err := ctx.Err(); err != nil {
		return domain.Order{}, fmt.Errorf("get order context: %w", err)
	}

	if id == uuid.Nil {
		return domain.Order{}, domain.ErrOrderIDRequired
	}

	if s.orderRepository == nil {
		return domain.Order{}, domain.ErrNotImplemented
	}

	order, err := s.orderRepository.GetOrder(ctx, id)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order with id='%s' from repository: %w", id, err)
	}

	return order, nil
}
