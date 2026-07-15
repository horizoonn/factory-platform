package assembly

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/assembly/internal/outbox"
	"github.com/horizoonn/factory-platform/assembly/internal/service/mocks"
)

func TestService_Assemble(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(context.Context, *mocks.Outbox, *mocks.ShipAssembledEncoder)
		wantErr    error
	}{
		{
			name: "success",
			setupMocks: func(ctx context.Context, outboxMock *mocks.Outbox, encoder *mocks.ShipAssembledEncoder) {
				encoded := encodedShipAssembledFixture()
				encoder.EXPECT().
					Encode(expectedShipAssembledEvent()).
					Return(encoded, nil).
					Once()
				encoded.SourceEventID = orderPaidEventID
				outboxMock.EXPECT().
					Enqueue(ctx, encoded).
					Return(true, nil).
					Once()
			},
		},
		{
			name: "duplicate OrderPaid event",
			setupMocks: func(ctx context.Context, outboxMock *mocks.Outbox, encoder *mocks.ShipAssembledEncoder) {
				encoded := encodedShipAssembledFixture()
				encoder.EXPECT().
					Encode(expectedShipAssembledEvent()).
					Return(encoded, nil).
					Once()
				encoded.SourceEventID = orderPaidEventID
				outboxMock.EXPECT().
					Enqueue(ctx, encoded).
					Return(false, nil).
					Once()
			},
		},
		{
			name: "error encode ShipAssembled event",
			setupMocks: func(_ context.Context, _ *mocks.Outbox, encoder *mocks.ShipAssembledEncoder) {
				encoder.EXPECT().
					Encode(expectedShipAssembledEvent()).
					Return(outbox.Event{}, errEncoder).
					Once()
			},
			wantErr: errEncoder,
		},
		{
			name: "error enqueue outbox event",
			setupMocks: func(ctx context.Context, outboxMock *mocks.Outbox, encoder *mocks.ShipAssembledEncoder) {
				encoded := encodedShipAssembledFixture()
				encoder.EXPECT().
					Encode(expectedShipAssembledEvent()).
					Return(encoded, nil).
					Once()
				encoded.SourceEventID = orderPaidEventID
				outboxMock.EXPECT().
					Enqueue(ctx, encoded).
					Return(false, errOutbox).
					Once()
			},
			wantErr: errOutbox,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			outboxMock := mocks.NewOutbox(t)
			encoder := mocks.NewShipAssembledEncoder(t)
			service := newTestService(outboxMock, encoder)
			tt.setupMocks(ctx, outboxMock, encoder)

			err := service.Assemble(ctx, orderPaidFixture())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_Assemble_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	service := newTestService(
		mocks.NewOutbox(t),
		mocks.NewShipAssembledEncoder(t),
	)

	err := service.Assemble(ctx, orderPaidFixture())
	require.ErrorIs(t, err, context.Canceled)
}
