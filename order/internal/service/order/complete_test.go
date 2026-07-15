package order

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/service/mocks"
)

func TestService_CompleteOrder(t *testing.T) {
	emptyEventIDEvent := validShipAssembledEvent()
	emptyEventIDEvent.ID = uuid.Nil

	emptyOrderIDEvent := validShipAssembledEvent()
	emptyOrderIDEvent.OrderID = uuid.Nil

	emptyUserIDEvent := validShipAssembledEvent()
	emptyUserIDEvent.UserID = uuid.Nil

	emptyOccurredAtEvent := validShipAssembledEvent()
	emptyOccurredAtEvent.OccurredAt = time.Time{}

	negativeBuildTimeEvent := validShipAssembledEvent()
	negativeBuildTimeEvent.BuildTimeSec = -1

	tests := []struct {
		name       string
		event      domain.ShipAssembledEvent
		setupMocks func(context.Context, *mocks.Repository, domain.ShipAssembledEvent)
		wantErr    error
	}{
		{
			name:  "success",
			event: validShipAssembledEvent(),
			setupMocks: func(ctx context.Context, repository *mocks.Repository, event domain.ShipAssembledEvent) {
				repository.EXPECT().
					CompleteOrder(ctx, event).
					Return(nil).
					Once()
			},
		},
		{
			name:  "error complete order in repository",
			event: validShipAssembledEvent(),
			setupMocks: func(ctx context.Context, repository *mocks.Repository, event domain.ShipAssembledEvent) {
				repository.EXPECT().
					CompleteOrder(ctx, event).
					Return(errRepository).
					Once()
			},
			wantErr: errRepository,
		},
		{
			name:    "error empty event id",
			event:   emptyEventIDEvent,
			wantErr: domain.ErrEventIDRequired,
		},
		{
			name:    "error empty order id",
			event:   emptyOrderIDEvent,
			wantErr: domain.ErrOrderIDRequired,
		},
		{
			name:    "error empty user id",
			event:   emptyUserIDEvent,
			wantErr: domain.ErrUserIDRequired,
		},
		{
			name:    "error empty occurred at",
			event:   emptyOccurredAtEvent,
			wantErr: domain.ErrOccurredAtRequired,
		},
		{
			name:    "error negative build time",
			event:   negativeBuildTimeEvent,
			wantErr: domain.ErrInvalidBuildTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repository := mocks.NewRepository(t)
			service := newServiceWithRepository(repository)

			if tt.setupMocks != nil {
				tt.setupMocks(ctx, repository, tt.event)
			}

			err := service.CompleteOrder(ctx, tt.event)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_CompleteOrder_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	service := newServiceWithRepository(mocks.NewRepository(t))
	err := service.CompleteOrder(ctx, validShipAssembledEvent())

	require.ErrorIs(t, err, context.Canceled)
}
