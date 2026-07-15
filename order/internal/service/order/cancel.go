package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
)

func (s *Service) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	if err := s.validateContext(ctx); err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}

	if orderID == uuid.Nil {
		return domain.ErrOrderIDRequired
	}

	if err := s.repository.CancelOrder(ctx, orderID); err != nil {
		return fmt.Errorf("cancel order in repository: %w", err)
	}

	return nil
}
