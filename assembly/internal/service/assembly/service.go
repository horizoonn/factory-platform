package assembly

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
	"github.com/horizoonn/factory-platform/assembly/internal/outbox"
)

const buildDuration = 10 * time.Second

type Outbox interface {
	Enqueue(ctx context.Context, event outbox.Event) (bool, error)
}

type ShipAssembledEncoder interface {
	Encode(event domain.ShipAssembledEvent) (outbox.Event, error)
}

type IDGenerator func() uuid.UUID

type Clock func() time.Time

type Service struct {
	outbox               Outbox
	shipAssembledEncoder ShipAssembledEncoder
	idGenerator          IDGenerator
	clock                Clock
}

func NewService(outbox Outbox, shipAssembledEncoder ShipAssembledEncoder) *Service {
	return &Service{
		outbox:               outbox,
		shipAssembledEncoder: shipAssembledEncoder,
		idGenerator:          uuid.New,
		clock:                time.Now,
	}
}
