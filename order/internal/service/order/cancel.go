package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
)

func (s *Service) CancelOrder(ctx context.Context, id uuid.UUID) error {
	if err := s.validateContext(ctx); err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}

	if id == uuid.Nil {
		return domain.ErrOrderIDRequired
	}

	order, err := s.repository.GetOrder(ctx, id)
	if err != nil {
		return fmt.Errorf("get order with id='%s' from repository: %w", id, err)
	}

	switch order.Status {
	case domain.OrderStatusPaid:
		return domain.ErrOrderAlreadyPaid
	case domain.OrderStatusCancelled:
		return nil
	case domain.OrderStatusPendingPayment:
	default:
		return domain.ErrInvalidOrderStatus
	}

	order.Status = domain.OrderStatusCancelled

	if _, err := s.repository.UpdateOrder(ctx, order); err != nil {
		return fmt.Errorf("update canceled order: %w", err)
	}

	return nil
}
