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
		setupMocks func(ctx context.Context, repository *mocks.Repository)
		wantErr    error
	}{
		{
			name:    "success pending payment",
			orderID: orderID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				order := validOrder()

				cancelledOrder := order
				cancelledOrder.Status = domain.OrderStatusCancelled

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()

				repository.EXPECT().
					UpdateOrder(ctx, cancelledOrder).
					Return(cancelledOrder, nil).
					Once()
			},
		},
		{
			name:    "success already cancelled",
			orderID: orderID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				order := validOrder()
				order.Status = domain.OrderStatusCancelled

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()
			},
		},
		{
			name:    "error empty order id",
			orderID: uuid.Nil,
			wantErr: domain.ErrOrderIDRequired,
		},
		{
			name:    "error order already paid",
			orderID: orderID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				order := validOrder()
				order.Status = domain.OrderStatusPaid

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()
			},
			wantErr: domain.ErrOrderAlreadyPaid,
		},
		{
			name:    "error invalid order status",
			orderID: orderID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				order := validOrder()
				order.Status = domain.OrderStatusUnknown

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()
			},
			wantErr: domain.ErrInvalidOrderStatus,
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
		{
			name:    "error update order in repository",
			orderID: orderID,
			setupMocks: func(ctx context.Context, repository *mocks.Repository) {
				order := validOrder()

				cancelledOrder := order
				cancelledOrder.Status = domain.OrderStatusCancelled

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()

				repository.EXPECT().
					UpdateOrder(ctx, cancelledOrder).
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

	repository := mocks.NewRepository(t)
	service := newServiceWithRepository(repository)

	err := service.CancelOrder(ctx, orderID)

	require.ErrorIs(t, err, context.Canceled)
}
