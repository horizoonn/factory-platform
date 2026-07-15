package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
)

func (s *Service) CompleteOrder(ctx context.Context, event domain.ShipAssembledEvent) error {
	if err := s.validateContext(ctx); err != nil {
		return fmt.Errorf("complete order: %w", err)
	}
	if event.ID == uuid.Nil {
		return domain.ErrEventIDRequired
	}
	if event.OrderID == uuid.Nil {
		return domain.ErrOrderIDRequired
	}
	if event.UserID == uuid.Nil {
		return domain.ErrUserIDRequired
	}
	if event.OccurredAt.IsZero() {
		return domain.ErrOccurredAtRequired
	}
	if event.BuildTimeSec < 0 {
		return domain.ErrInvalidBuildTime
	}

	if err := s.repository.CompleteOrder(ctx, event); err != nil {
		return fmt.Errorf("complete order in repository: %w", err)
	}

	return nil
}
