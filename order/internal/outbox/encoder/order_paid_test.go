package encoder

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	eventsv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/events/v1"
)

func TestOrderPaid_Encode(t *testing.T) {
	const topic = "order.paid.v1"

	event := domain.OrderPaidEvent{
		ID:            uuid.New(),
		OrderID:       uuid.New(),
		UserID:        uuid.New(),
		PaymentMethod: domain.PaymentMethodCard,
		TransactionID: uuid.New(),
		OccurredAt:    time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC),
	}

	encoded, err := NewOrderPaid(topic).Encode(event)
	require.NoError(t, err)
	assert.Equal(t, event.ID, encoded.ID)
	assert.Equal(t, event.OrderID, encoded.AggregateID)
	assert.Equal(t, orderPaidType, encoded.Type)
	assert.Equal(t, topic, encoded.Topic)
	assert.Equal(t, []byte(event.OrderID.String()), encoded.Key)
	assert.Equal(t, map[string]string{"event-type": orderPaidType}, encoded.Headers)
	assert.Equal(t, event.OccurredAt, encoded.AvailableAt)

	var message eventsv1.OrderPaid
	require.NoError(t, proto.Unmarshal(encoded.Payload, &message))
	assert.Equal(t, event.ID.String(), message.GetEventUuid())
	assert.Equal(t, event.OrderID.String(), message.GetOrderUuid())
	assert.Equal(t, event.UserID.String(), message.GetUserUuid())
	assert.Equal(t, string(event.PaymentMethod), message.GetPaymentMethod())
	assert.Equal(t, event.TransactionID.String(), message.GetTransactionUuid())
	require.NotNil(t, message.GetOccurredAt())
	assert.Equal(t, event.OccurredAt, message.GetOccurredAt().AsTime())
}

func TestOrderPaid_Encode_InvalidTimestamp(t *testing.T) {
	event := domain.OrderPaidEvent{
		ID:            uuid.New(),
		OrderID:       uuid.New(),
		UserID:        uuid.New(),
		PaymentMethod: domain.PaymentMethodCard,
		TransactionID: uuid.New(),
		OccurredAt:    time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := NewOrderPaid("order.paid.v1").Encode(event)
	require.Error(t, err)
}
