package assembly

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
	"github.com/horizoonn/factory-platform/assembly/internal/service/mocks"
)

func TestServiceAssemble(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	outbox := mocks.NewShipAssembledOutbox(t)
	service := newTestService(outbox)

	outbox.EXPECT().
		EnqueueShipAssembled(ctx, orderPaidEventID, expectedShipAssembledEvent()).
		Return(true, nil).
		Once()

	err := service.Assemble(ctx, orderPaidFixture())

	require.NoError(t, err)
}

func TestServiceAssembleDuplicateOrderPaid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	outbox := mocks.NewShipAssembledOutbox(t)
	service := newTestService(outbox)

	outbox.EXPECT().
		EnqueueShipAssembled(ctx, orderPaidEventID, expectedShipAssembledEvent()).
		Return(false, nil).
		Once()

	err := service.Assemble(ctx, orderPaidFixture())

	require.NoError(t, err)
}

func TestServiceAssembleOutboxError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	errOutbox := errors.New("outbox unavailable")
	outbox := mocks.NewShipAssembledOutbox(t)
	service := newTestService(outbox)

	outbox.EXPECT().
		EnqueueShipAssembled(ctx, orderPaidEventID, expectedShipAssembledEvent()).
		Return(false, errOutbox).
		Once()

	err := service.Assemble(ctx, orderPaidFixture())

	require.Error(t, err)
	assert.ErrorIs(t, err, errOutbox)
}

func TestServiceAssembleCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	outbox := mocks.NewShipAssembledOutbox(t)
	service := newTestService(outbox)

	err := service.Assemble(ctx, orderPaidFixture())

	require.ErrorIs(t, err, context.Canceled)
}

var (
	orderPaidEventID  = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f985001")
	orderID           = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f985002")
	userID            = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f985003")
	transactionID     = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f985004")
	shipAssembledID   = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f985005")
	assemblyStartedAt = time.Date(2026, time.July, 11, 9, 0, 0, 0, time.UTC)
)

func newTestService(outbox ShipAssembledOutbox) *Service {
	service := NewService(outbox)
	service.idGenerator = func() uuid.UUID { return shipAssembledID }
	service.clock = func() time.Time { return assemblyStartedAt }

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
