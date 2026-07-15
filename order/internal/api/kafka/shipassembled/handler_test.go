package shipassembled

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/service/mocks"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer"
	eventsv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/events/v1"
)

func TestHandler_Handle(t *testing.T) {
	ctx := context.Background()
	eventID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	orderID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	occurredAt := time.Date(2026, time.July, 14, 12, 0, 0, 0, time.UTC)
	expectedEvent := domain.ShipAssembledEvent{
		ID:           eventID,
		OrderID:      orderID,
		UserID:       userID,
		BuildTimeSec: 42,
		OccurredAt:   occurredAt,
	}

	payload, err := proto.Marshal(&eventsv1.ShipAssembled{
		EventUuid:    eventID.String(),
		OrderUuid:    orderID.String(),
		UserUuid:     userID.String(),
		BuildTimeSec: 42,
		OccurredAt:   timestamppb.New(occurredAt),
	})
	require.NoError(t, err)

	validRecord := kafka.Record{Message: kafka.Message{Value: payload}}
	errRepository := errors.New("database unavailable")

	tests := []struct {
		name          string
		record        kafka.Record
		setupMocks    func(context.Context, *mocks.OrderService)
		wantErr       bool
		wantErrIs     error
		wantPermanent bool
	}{
		{
			name:   "success",
			record: validRecord,
			setupMocks: func(ctx context.Context, service *mocks.OrderService) {
				service.EXPECT().
					CompleteOrder(ctx, expectedEvent).
					Return(nil).
					Once()
			},
		},
		{
			name:          "error malformed message is permanent",
			record:        kafka.Record{Message: kafka.Message{Value: []byte{0xff}}},
			wantErr:       true,
			wantPermanent: true,
		},
		{
			name:   "error domain conflict is permanent",
			record: validRecord,
			setupMocks: func(ctx context.Context, service *mocks.OrderService) {
				service.EXPECT().
					CompleteOrder(ctx, expectedEvent).
					Return(domain.ErrOrderUserMismatch).
					Once()
			},
			wantErr:       true,
			wantErrIs:     domain.ErrOrderUserMismatch,
			wantPermanent: true,
		},
		{
			name:   "error infrastructure failure is retryable",
			record: validRecord,
			setupMocks: func(ctx context.Context, service *mocks.OrderService) {
				service.EXPECT().
					CompleteOrder(ctx, expectedEvent).
					Return(errRepository).
					Once()
			},
			wantErr:   true,
			wantErrIs: errRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := mocks.NewOrderService(t)
			handler := NewHandler(service)

			if tt.setupMocks != nil {
				tt.setupMocks(ctx, service)
			}

			err := handler.Handle(ctx, tt.record)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.wantPermanent, consumer.IsPermanent(err))
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}
