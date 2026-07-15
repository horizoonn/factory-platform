//go:build integration

package integration

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	orderrepo "github.com/horizoonn/factory-platform/order/internal/repository/order"
	outboxrepo "github.com/horizoonn/factory-platform/order/internal/repository/outbox"
)

func TestOrderRepository_CreateOrder(t *testing.T) {
	testEnv.truncateOrders(t)

	repository := newOrderRepository()
	expected := baseOrder()

	actual, err := repository.CreateOrder(testContext(t), expected)

	require.NoError(t, err)
	assertOrder(t, expected, actual)
	assert.Nil(t, actual.TransactionID)
	assert.Nil(t, actual.PaymentMethod)
	assert.False(t, actual.CreatedAt.IsZero())
	assert.False(t, actual.UpdatedAt.IsZero())
}

func TestOrderRepository_GetOrder(t *testing.T) {
	testEnv.truncateOrders(t)

	repository := newOrderRepository()
	expected := insertOrder(t, baseOrder())

	actual, err := repository.GetOrder(testContext(t), expected.ID)

	require.NoError(t, err)
	assertOrder(t, expected, actual)
	assert.True(t, expected.CreatedAt.Equal(actual.CreatedAt))
	assert.True(t, expected.UpdatedAt.Equal(actual.UpdatedAt))
}

func TestOrderRepository_GetOrder_NullablePaymentFields(t *testing.T) {
	testEnv.truncateOrders(t)

	repository := newOrderRepository()
	expected := baseOrder()
	expected.TransactionID = nil
	expected.PaymentMethod = nil
	expected.Status = domain.OrderStatusPendingPayment
	insertOrder(t, expected)

	actual, err := repository.GetOrder(testContext(t), expected.ID)

	require.NoError(t, err)
	assert.Nil(t, actual.TransactionID)
	assert.Nil(t, actual.PaymentMethod)
	assert.Equal(t, domain.OrderStatusPendingPayment, actual.Status)
}

func TestOrderRepository_GetOrder_NotFound(t *testing.T) {
	testEnv.truncateOrders(t)

	repository := newOrderRepository()

	_, err := repository.GetOrder(testContext(t), uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982404"))

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrNotFound))
}

func TestOrderRepository_CancelOrder(t *testing.T) {
	testEnv.truncateOrders(t)

	repository := newOrderRepository()
	existing := insertOrder(t, baseOrder())
	err := repository.CancelOrder(testContext(t), existing.ID)

	require.NoError(t, err)
	assert.Equal(t, domain.OrderStatusCancelled, orderStatus(t, existing.ID))
}

func TestOrderRepository_CancelOrder_Idempotent(t *testing.T) {
	testEnv.truncateOrders(t)

	repository := newOrderRepository()
	order := baseOrder()
	order.Status = domain.OrderStatusCancelled
	insertOrder(t, order)

	err := repository.CancelOrder(testContext(t), order.ID)

	require.NoError(t, err)
	assert.Equal(t, domain.OrderStatusCancelled, orderStatus(t, order.ID))
}

func TestOrderRepository_CancelOrder_DoesNotCancelPaidOrder(t *testing.T) {
	testEnv.truncateOrders(t)

	repository := newOrderRepository()
	order := baseOrder()
	transactionID := uuid.New()
	paymentMethod := domain.PaymentMethodCard
	order.TransactionID = &transactionID
	order.PaymentMethod = &paymentMethod
	order.Status = domain.OrderStatusPaid
	insertOrder(t, order)

	err := repository.CancelOrder(testContext(t), order.ID)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrOrderAlreadyPaid)
	assert.Equal(t, domain.OrderStatusPaid, orderStatus(t, order.ID))
}

func TestOrderRepository_CancelOrder_NotFound(t *testing.T) {
	testEnv.truncateOrders(t)

	repository := newOrderRepository()
	err := repository.CancelOrder(testContext(t), baseOrder().ID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrNotFound))
}

func TestOrderRepository_CreateOrder_RejectsEmptyPartIDs(t *testing.T) {
	testEnv.truncateOrders(t)

	repository := newOrderRepository()
	order := baseOrder()
	order.PartIDs = nil

	_, err := repository.CreateOrder(testContext(t), order)

	require.Error(t, err)
}

func baseOrder() domain.Order {
	return domain.Order{
		ID:     uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982001"),
		UserID: uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982101"),
		PartIDs: []uuid.UUID{
			uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982201"),
			uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982202"),
		},
		TotalPrice: 199.95,
		Status:     domain.OrderStatusPendingPayment,
	}
}

func insertOrder(t *testing.T, order domain.Order) domain.Order {
	t.Helper()

	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	updatedAt := createdAt

	_, err := testEnv.pool.Exec(
		testContext(t), `
		INSERT INTO platform.orders (
			id, user_id, part_ids, total_price,
			transaction_id, payment_method, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
		order.ID,
		order.UserID,
		order.PartIDs,
		order.TotalPrice,
		order.TransactionID,
		nullablePaymentMethod(order.PaymentMethod),
		string(order.Status),
		createdAt,
		updatedAt,
	)
	require.NoError(t, err)

	order.CreatedAt = createdAt
	order.UpdatedAt = updatedAt

	return order
}

func assertOrder(t *testing.T, expected, actual domain.Order) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.UserID, actual.UserID)
	assert.ElementsMatch(t, expected.PartIDs, actual.PartIDs)
	assert.InEpsilon(t, expected.TotalPrice, actual.TotalPrice, 0.0001)
	assert.Equal(t, expected.Status, actual.Status)
}

func nullablePaymentMethod(method *domain.PaymentMethod) sql.NullString {
	if method == nil {
		return sql.NullString{}
	}

	return sql.NullString{String: string(*method), Valid: true}
}

func newOrderRepository() *orderrepo.Repository {
	return orderrepo.NewRepository(
		testEnv.pool,
		outboxrepo.NewRepository(testEnv.pool),
	)
}
