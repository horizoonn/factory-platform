package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
)

func (s *OrderService) CancelOrder(ctx context.Context, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("cancel order context: %w", err)
	}

	if id == uuid.Nil {
		return domain.ErrOrderIDRequired
	}

	if s.orderRepository == nil {
		return domain.ErrNotImplemented
	}

	order, err := s.orderRepository.GetOrder(ctx, id)
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

	if _, err := s.orderRepository.UpdateOrder(ctx, order); err != nil {
		return fmt.Errorf("update canceled order: %w", err)
	}

	return nil
}
