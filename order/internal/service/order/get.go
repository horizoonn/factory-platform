package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
)

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (domain.Order, error) {
	if err := s.validateContext(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}

	if id == uuid.Nil {
		return domain.Order{}, domain.ErrOrderIDRequired
	}

	order, err := s.repository.GetOrder(ctx, id)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order with id='%s' from repository: %w", id, err)
	}

	return order, nil
}
