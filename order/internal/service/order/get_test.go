package order

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/service/mocks"
)

func TestService_GetOrder(t *testing.T) {
	tests := []struct {
		name       string
		orderID    uuid.UUID
		setupMocks func(ctx context.Context, repository *mocks.Repository)
		want       domain.Order
		wantErr    error
	}{
		{
			name:    "success",
			orderID: orderID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				order := validOrder()

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()
			},
			want: validOrder(),
		},
		{
			name:    "error empty order id",
			orderID: uuid.Nil,
			wantErr: domain.ErrOrderIDRequired,
		},
		{
			name:    "error get order from repository",
			orderID: orderID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(domain.Order{}, errRepository).
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

			got, err := service.GetOrder(ctx, tt.orderID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, domain.Order{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_GetOrder_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repository := mocks.NewRepository(t)
	service := newServiceWithRepository(repository)

	got, err := service.GetOrder(ctx, orderID)

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, domain.Order{}, got)
}
