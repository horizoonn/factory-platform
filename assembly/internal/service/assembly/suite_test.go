package assembly

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
	"github.com/horizoonn/factory-platform/assembly/internal/outbox"
)

var (
	orderPaidEventID  = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f985001")
	orderID           = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f985002")
	userID            = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f985003")
	transactionID     = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f985004")
	shipAssembledID   = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f985005")
	assemblyStartedAt = time.Date(2026, time.July, 11, 9, 0, 0, 0, time.UTC)
	errEncoder        = errors.New("encoder error")
	errOutbox         = errors.New("outbox error")
)

func newTestService(outboxMock Outbox, encoder ShipAssembledEncoder) *Service {
	service := NewService(outboxMock, encoder)
	service.idGenerator = func() uuid.UUID {
		return shipAssembledID
	}
	service.clock = func() time.Time {
		return assemblyStartedAt
	}

	return service
}

func orderPaidFixture() domain.OrderPaidEvent {
	return domain.OrderPaidEvent{
		ID:            orderPaidEventID,
		OrderID:       orderID,
		UserID:        userID,
		PaymentMethod: "CARD",
		TransactionID: transactionID,
		OccurredAt:    assemblyStartedAt.Add(-time.Second),
	}
}

func expectedShipAssembledEvent() domain.ShipAssembledEvent {
	return domain.ShipAssembledEvent{
		ID:           shipAssembledID,
		OrderID:      orderID,
		UserID:       userID,
		BuildTimeSec: int64(buildDuration.Seconds()),
		OccurredAt:   assemblyStartedAt.Add(buildDuration),
	}
}

func encodedShipAssembledFixture() outbox.Event {
	return outbox.Event{
		ID:          shipAssembledID,
		AggregateID: orderID,
		Topic:       "assembly.ship-assembled.v1",
		Key:         []byte(orderID.String()),
		Payload:     []byte("payload"),
		Headers:     map[string]string{"event-type": "events.v1.ShipAssembled"},
		AvailableAt: assemblyStartedAt.Add(buildDuration),
	}
}
