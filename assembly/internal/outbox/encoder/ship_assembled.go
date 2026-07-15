package encoder

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
	"github.com/horizoonn/factory-platform/assembly/internal/outbox"
	eventsv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/events/v1"
)

const (
	shipAssembledTopic = "assembly.ship-assembled.v1"
	shipAssembledType  = "events.v1.ShipAssembled" //nolint:gosec // G101: Kafka event type, not a credential.
)

type ShipAssembled struct{}

func NewShipAssembled() *ShipAssembled {
	return &ShipAssembled{}
}

func (e *ShipAssembled) Encode(event domain.ShipAssembledEvent) (outbox.Event, error) {
	occurredAt := timestamppb.New(event.OccurredAt)
	if err := occurredAt.CheckValid(); err != nil {
		return outbox.Event{}, fmt.Errorf("validate ShipAssembled occurred_at: %w", err)
	}

	payload, err := proto.Marshal(&eventsv1.ShipAssembled{
		EventUuid:    event.ID.String(),
		OrderUuid:    event.OrderID.String(),
		UserUuid:     event.UserID.String(),
		BuildTimeSec: event.BuildTimeSec,
		OccurredAt:   occurredAt,
	})
	if err != nil {
		return outbox.Event{}, fmt.Errorf("marshal ShipAssembled protobuf: %w", err)
	}

	return outbox.Event{
		ID:          event.ID,
		AggregateID: event.OrderID,
		Topic:       shipAssembledTopic,
		Key:         []byte(event.OrderID.String()),
		Payload:     payload,
		Headers:     map[string]string{"event-type": shipAssembledType},
		AvailableAt: event.OccurredAt,
	}, nil
}
