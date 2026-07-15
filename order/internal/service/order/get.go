package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
)

func (s *Service) GetOrder(ctx context.Context, orderID uuid.UUID) (domain.Order, error) {
	if err := s.validateContext(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}

	if orderID == uuid.Nil {
		return domain.Order{}, domain.ErrOrderIDRequired
	}

	order, err := s.repository.GetOrder(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order id=%s from repository: %w", orderID, err)
	}

	return order, nil
}
