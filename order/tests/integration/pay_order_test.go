//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/outbox"
)

func TestOrderRepository_MarkPaidAndEnqueueOrderPaid(t *testing.T) {
	testEnv.truncateOutbox(t)
	testEnv.truncateOrders(t)

	order := insertOrder(t, baseOrder())
	transactionID := uuid.New()
	paymentMethod := domain.PaymentMethodCard
	order.TransactionID = &transactionID
	order.PaymentMethod = &paymentMethod
	order.Status = domain.OrderStatusPaid
	event := orderPaidOutboxFixture(order.ID)

	updated, err := newOrderRepository().MarkPaidAndEnqueueOrderPaid(
		testContext(t),
		order,
		event,
	)

	require.NoError(t, err)
	assert.Equal(t, domain.OrderStatusPaid, updated.Status)
	require.NotNil(t, updated.TransactionID)
	assert.Equal(t, transactionID, *updated.TransactionID)
	assert.EqualValues(t, 1, outboxEventCount(t, event.ID))
}

func TestOrderRepository_MarkPaidAndEnqueueOrderPaid_RollsBackForCancelledOrder(t *testing.T) {
	testEnv.truncateOutbox(t)
	testEnv.truncateOrders(t)

	order := baseOrder()
	order.Status = domain.OrderStatusCancelled
	insertOrder(t, order)
	transactionID := uuid.New()
	paymentMethod := domain.PaymentMethodCard
	order.TransactionID = &transactionID
	order.PaymentMethod = &paymentMethod
	order.Status = domain.OrderStatusPaid
	event := orderPaidOutboxFixture(order.ID)

	_, err := newOrderRepository().MarkPaidAndEnqueueOrderPaid(testContext(t), order, event)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrOrderCancelled)
	assert.Equal(t, domain.OrderStatusCancelled, orderStatus(t, order.ID))
	assert.EqualValues(t, 0, outboxEventCount(t, event.ID))
}

func orderPaidOutboxFixture(orderID uuid.UUID) outbox.Event {
	eventID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	return outbox.Event{
		ID:          eventID,
		AggregateID: orderID,
		Type:        "events.v1.OrderPaid",
		Topic:       "order.paid.v1",
		Key:         []byte(orderID.String()),
		Payload:     []byte("order paid payload"),
		Headers:     map[string]string{"event-type": "events.v1.OrderPaid"},
		AvailableAt: now,
	}
}

func outboxEventCount(t *testing.T, eventID uuid.UUID) int64 {
	t.Helper()

	var count int64
	err := testEnv.pool.QueryRow(
		testContext(t),
		"SELECT COUNT(*) FROM platform.order_outbox_events WHERE id = $1",
		eventID,
	).Scan(&count)
	require.NoError(t, err)

	return count
}
