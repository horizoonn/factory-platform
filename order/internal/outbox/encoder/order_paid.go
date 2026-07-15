package encoder

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/outbox"
	eventsv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/events/v1"
)

const (
	orderPaidType = "events.v1.OrderPaid"
)

type OrderPaid struct {
	topic string
}

func NewOrderPaid(topic string) *OrderPaid {
	return &OrderPaid{topic: topic}
}

func (e *OrderPaid) Encode(event domain.OrderPaidEvent) (outbox.Event, error) {
	occurredAt := timestamppb.New(event.OccurredAt)
	if err := occurredAt.CheckValid(); err != nil {
		return outbox.Event{}, fmt.Errorf("validate OrderPaid occurred_at: %w", err)
	}

	payload, err := proto.Marshal(&eventsv1.OrderPaid{
		EventUuid:       event.ID.String(),
		OrderUuid:       event.OrderID.String(),
		UserUuid:        event.UserID.String(),
		PaymentMethod:   string(event.PaymentMethod),
		TransactionUuid: event.TransactionID.String(),
		OccurredAt:      occurredAt,
	})
	if err != nil {
		return outbox.Event{}, fmt.Errorf("marshal OrderPaid protobuf: %w", err)
	}

	return outbox.Event{
		ID:          event.ID,
		AggregateID: event.OrderID,
		Type:        orderPaidType,
		Topic:       e.topic,
		Key:         []byte(event.OrderID.String()),
		Payload:     payload,
		Headers:     map[string]string{"event-type": orderPaidType},
		AvailableAt: event.OccurredAt,
	}, nil
}
