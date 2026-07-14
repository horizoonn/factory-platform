package outbox

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
	eventsv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/events/v1"
)

const shipAssembledTopic = "assembly.ship-assembled.v1"

type Repository interface {
	Enqueue(ctx context.Context, event Event) (bool, error)
}

type ShipAssembledWriter struct {
	repository Repository
}

func NewShipAssembledWriter(repository Repository) *ShipAssembledWriter {
	return &ShipAssembledWriter{
		repository: repository,
	}
}

func (w *ShipAssembledWriter) EnqueueShipAssembled(
	ctx context.Context,
	sourceEventID uuid.UUID,
	event domain.ShipAssembledEvent,
) (bool, error) {
	occurredAt := timestamppb.New(event.OccurredAt)
	if err := occurredAt.CheckValid(); err != nil {
		return false, fmt.Errorf("validate ShipAssembled occurred_at: %w", err)
	}

	payload, err := proto.Marshal(&eventsv1.ShipAssembled{
		EventUuid:    event.ID.String(),
		OrderUuid:    event.OrderID.String(),
		UserUuid:     event.UserID.String(),
		BuildTimeSec: event.BuildTimeSec,
		OccurredAt:   occurredAt,
	})
	if err != nil {
		return false, fmt.Errorf("marshal ShipAssembled event: %w", err)
	}

	created, err := w.repository.Enqueue(ctx, Event{
		ID:            event.ID,
		SourceEventID: sourceEventID,
		AggregateID:   event.OrderID,
		Topic:         shipAssembledTopic,
		Key:           []byte(event.OrderID.String()),
		Payload:       payload,
		Headers:       map[string]string{"event-type": "events.v1.ShipAssembled"},
		AvailableAt:   event.OccurredAt,
	})
	if err != nil {
		return false, fmt.Errorf("enqueue ShipAssembled outbox event: %w", err)
	}

	return created, nil
}
