package shipassembled

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
	eventsv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/events/v1"
)

func recordToShipAssembledEvent(record kafka.Record) (domain.ShipAssembledEvent, error) {
	var msg eventsv1.ShipAssembled

	if err := proto.Unmarshal(record.Value, &msg); err != nil {
		return domain.ShipAssembledEvent{}, fmt.Errorf("unmarshal ShipAssembled event: %w", err)
	}

	occurredAt := msg.GetOccurredAt()
	if occurredAt == nil {
		return domain.ShipAssembledEvent{}, errors.New("validate ShipAssembled occurred_at: timestamp is required")
	}
	if err := occurredAt.CheckValid(); err != nil {
		return domain.ShipAssembledEvent{}, fmt.Errorf("validate ShipAssembled occurred_at: %w", err)
	}

	eventID, err := parseShipAssembledUUID("event_uuid", msg.GetEventUuid())
	if err != nil {
		return domain.ShipAssembledEvent{}, err
	}

	orderID, err := parseShipAssembledUUID("order_uuid", msg.GetOrderUuid())
	if err != nil {
		return domain.ShipAssembledEvent{}, err
	}

	userID, err := parseShipAssembledUUID("user_uuid", msg.GetUserUuid())
	if err != nil {
		return domain.ShipAssembledEvent{}, err
	}

	if msg.GetBuildTimeSec() < 0 {
		return domain.ShipAssembledEvent{}, errors.New(
			"validate ShipAssembled build_time_sec: must not be negative",
		)
	}

	return domain.ShipAssembledEvent{
		ID:           eventID,
		OrderID:      orderID,
		UserID:       userID,
		BuildTimeSec: msg.GetBuildTimeSec(),
		OccurredAt:   occurredAt.AsTime(),
	}, nil
}

func parseShipAssembledUUID(field, value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse ShipAssembled %s: %w", field, err)
	}
	if id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("parse ShipAssembled %s: UUID is nil", field)
	}

	return id, nil
}
