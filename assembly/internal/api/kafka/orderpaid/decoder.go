package orderpaid

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
	eventsv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/events/v1"
)

func recordToOrderPaidEvent(record kafka.Record) (domain.OrderPaidEvent, error) {
	var msg eventsv1.OrderPaid

	if err := proto.Unmarshal(record.Value, &msg); err != nil {
		return domain.OrderPaidEvent{}, fmt.Errorf("unmarshal OrderPaid event: %w", err)
	}

	if err := msg.GetOccurredAt().CheckValid(); err != nil {
		return domain.OrderPaidEvent{}, fmt.Errorf("validate OrderPaid occurred_at: %w", err)
	}

	eventID, err := parseOrderPaidUUID("event_uuid", msg.GetEventUuid())
	if err != nil {
		return domain.OrderPaidEvent{}, err
	}

	orderID, err := parseOrderPaidUUID("order_uuid", msg.GetOrderUuid())
	if err != nil {
		return domain.OrderPaidEvent{}, err
	}

	userID, err := parseOrderPaidUUID("user_uuid", msg.GetUserUuid())
	if err != nil {
		return domain.OrderPaidEvent{}, err
	}

	transactionID, err := parseOrderPaidUUID("transaction_uuid", msg.GetTransactionUuid())
	if err != nil {
		return domain.OrderPaidEvent{}, err
	}

	return domain.OrderPaidEvent{
		ID:            eventID,
		OrderID:       orderID,
		UserID:        userID,
		PaymentMethod: msg.GetPaymentMethod(),
		TransactionID: transactionID,
		OccurredAt:    msg.GetOccurredAt().AsTime(),
	}, nil
}

func parseOrderPaidUUID(field, value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse OrderPaid %s: %w", field, err)
	}
	if id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("parse OrderPaid %s: UUID is nil", field)
	}

	return id, nil
}
