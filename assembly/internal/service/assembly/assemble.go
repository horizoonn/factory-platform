package assembly

import (
	"context"
	"fmt"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
)

func (s *Service) Assemble(ctx context.Context, event domain.OrderPaidEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	availableAt := s.clock().UTC().Add(buildDuration)
	shipAssembled := domain.ShipAssembledEvent{
		ID:           s.idGenerator(),
		OrderID:      event.OrderID,
		UserID:       event.UserID,
		BuildTimeSec: int64(buildDuration.Seconds()),
		OccurredAt:   availableAt,
	}

	if _, err := s.outbox.EnqueueShipAssembled(ctx, event.ID, shipAssembled); err != nil {
		return fmt.Errorf("schedule ShipAssembled event: %w", err)
	}

	return nil
}
