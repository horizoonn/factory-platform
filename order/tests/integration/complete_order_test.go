//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/order/internal/domain"
)

func TestOrderRepository_CompleteOrder(t *testing.T) {
	testEnv.truncateInbox(t)
	testEnv.truncateOrders(t)

	order := paidOrderFixture(t)
	event := shipAssembledEvent(order)

	err := newOrderRepository().CompleteOrder(testContext(t), event)
	require.NoError(t, err)

	assert.Equal(t, domain.OrderStatusCompleted, orderStatus(t, order.ID))
	assert.EqualValues(t, 1, inboxEventCount(t, event.ID))
}

func TestOrderRepository_CompleteOrder_DuplicateEvent(t *testing.T) {
	testEnv.truncateInbox(t)
	testEnv.truncateOrders(t)

	order := paidOrderFixture(t)
	event := shipAssembledEvent(order)
	repository := newOrderRepository()
	require.NoError(t, repository.CompleteOrder(testContext(t), event))

	err := repository.CompleteOrder(testContext(t), event)
	require.NoError(t, err)
	assert.Equal(t, domain.OrderStatusCompleted, orderStatus(t, order.ID))
	assert.EqualValues(t, 1, inboxEventCount(t, event.ID))
}

func TestOrderRepository_CompleteOrder_RollsBackInboxOnNotFound(t *testing.T) {
	testEnv.truncateInbox(t)
	testEnv.truncateOrders(t)

	event := shipAssembledEvent(baseOrder())
	err := newOrderRepository().CompleteOrder(testContext(t), event)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.EqualValues(t, 0, inboxEventCount(t, event.ID))
}

func TestOrderRepository_CompleteOrder_RollsBackInboxOnUserMismatch(t *testing.T) {
	testEnv.truncateInbox(t)
	testEnv.truncateOrders(t)

	order := paidOrderFixture(t)
	event := shipAssembledEvent(order)
	event.UserID = uuid.New()

	err := newOrderRepository().CompleteOrder(testContext(t), event)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrOrderUserMismatch)
	assert.Equal(t, domain.OrderStatusPaid, orderStatus(t, order.ID))
	assert.EqualValues(t, 0, inboxEventCount(t, event.ID))
}

func TestOrderRepository_CompleteOrder_RollsBackInboxForInvalidStatus(t *testing.T) {
	testEnv.truncateInbox(t)
	testEnv.truncateOrders(t)

	order := insertOrder(t, baseOrder())
	event := shipAssembledEvent(order)

	err := newOrderRepository().CompleteOrder(testContext(t), event)
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidOrderStatus))
	assert.Equal(t, domain.OrderStatusPendingPayment, orderStatus(t, order.ID))
	assert.EqualValues(t, 0, inboxEventCount(t, event.ID))
}

func TestOrderRepository_CompleteOrder_AlreadyCompletedWithNewEvent(t *testing.T) {
	testEnv.truncateInbox(t)
	testEnv.truncateOrders(t)

	order := paidOrderFixture(t)
	order.Status = domain.OrderStatusCompleted
	_, err := testEnv.pool.Exec(
		testContext(t),
		"UPDATE platform.orders SET status = $1 WHERE id = $2",
		order.Status,
		order.ID,
	)
	require.NoError(t, err)

	event := shipAssembledEvent(order)
	err = newOrderRepository().CompleteOrder(testContext(t), event)
	require.NoError(t, err)
	assert.EqualValues(t, 1, inboxEventCount(t, event.ID))
}

func paidOrderFixture(t *testing.T) domain.Order {
	t.Helper()

	order := baseOrder()
	transactionID := uuid.New()
	paymentMethod := domain.PaymentMethodCard
	order.TransactionID = &transactionID
	order.PaymentMethod = &paymentMethod
	order.Status = domain.OrderStatusPaid

	return insertOrder(t, order)
}

func shipAssembledEvent(order domain.Order) domain.ShipAssembledEvent {
	return domain.ShipAssembledEvent{
		ID:           uuid.New(),
		OrderID:      order.ID,
		UserID:       order.UserID,
		BuildTimeSec: 42,
		OccurredAt:   time.Now().UTC(),
	}
}

func orderStatus(t *testing.T, orderID uuid.UUID) domain.OrderStatus {
	t.Helper()

	var status domain.OrderStatus
	err := testEnv.pool.QueryRow(
		testContext(t),
		"SELECT status FROM platform.orders WHERE id = $1",
		orderID,
	).Scan(&status)
	require.NoError(t, err)

	return status
}

func inboxEventCount(t *testing.T, eventID uuid.UUID) int64 {
	t.Helper()

	var count int64
	err := testEnv.pool.QueryRow(
		testContext(t),
		"SELECT COUNT(*) FROM platform.order_inbox_events WHERE event_id = $1",
		eventID,
	).Scan(&count)
	require.NoError(t, err)

	return count
}
