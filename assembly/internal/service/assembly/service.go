package assembly

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
)

const buildDuration = 10 * time.Second

type ShipAssembledOutbox interface {
	EnqueueShipAssembled(
		ctx context.Context,
		sourceEventID uuid.UUID,
		event domain.ShipAssembledEvent,
	) (bool, error)
}

type IDGenerator func() uuid.UUID

type Clock func() time.Time

type Service struct {
	outbox      ShipAssembledOutbox
	idGenerator IDGenerator
	clock       Clock
}

func NewService(outbox ShipAssembledOutbox) *Service {
	return &Service{
		outbox:      outbox,
		idGenerator: uuid.New,
		clock:       time.Now,
	}
}
