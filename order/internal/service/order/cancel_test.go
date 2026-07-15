package order

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/service/mocks"
)

func TestService_CancelOrder(t *testing.T) {
	tests := []struct {
		name       string
		orderID    uuid.UUID
		setupMocks func(context.Context, *mocks.Repository)
		wantErr    error
	}{
		{
			name:    "success",
			orderID: orderID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				repository.EXPECT().
					CancelOrder(ctx, orderID).
					Return(nil).
					Once()
			},
		},
		{
			name:    "error empty order id",
			orderID: uuid.Nil,
			wantErr: domain.ErrOrderIDRequired,
		},
		{
			name:    "error cancel order in repository",
			orderID: orderID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				repository.EXPECT().
					CancelOrder(ctx, orderID).
					Return(errRepository).
					Once()
			},
			wantErr: errRepository,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repository := mocks.NewRepository(t)
			service := newServiceWithRepository(repository)

			if tt.setupMocks != nil {
				tt.setupMocks(ctx, repository)
			}

			err := service.CancelOrder(ctx, tt.orderID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_CancelOrder_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	service := newServiceWithRepository(mocks.NewRepository(t))
	err := service.CancelOrder(ctx, orderID)

	require.ErrorIs(t, err, context.Canceled)
}
