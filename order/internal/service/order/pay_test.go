package order

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clientdto "github.com/horizoonn/factory-platform/order/internal/client/dto"
	"github.com/horizoonn/factory-platform/order/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
	"github.com/horizoonn/factory-platform/order/internal/service/mocks"
)

func TestService_PayOrder(t *testing.T) {
	tests := []struct {
		name       string
		req        servicedto.PayOrderRequest
		setupMocks func(ctx context.Context, repository *mocks.Repository, paymentClient *mocks.PaymentClient)
		want       domain.Order
		wantErr    error
	}{
		{
			name: "success",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			setupMocks: func(ctx context.Context, repository *mocks.Repository, paymentClient *mocks.PaymentClient) {
				order := validOrder()
				expectedOrder := paidOrder(domain.PaymentMethodCard)

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()

				paymentClient.EXPECT().
					PayOrder(ctx, clientdto.PayOrderRequest{
						OrderID:       orderID,
						UserID:        userID,
						PaymentMethod: domain.PaymentMethodCard,
					}).
					Return(clientdto.PayOrderResponse{TransactionID: transactionID}, nil).
					Once()

				repository.EXPECT().
					UpdateOrder(ctx, expectedOrder).
					Return(expectedOrder, nil).
					Once()
			},
			want: paidOrder(domain.PaymentMethodCard),
		},
		{
			name: "error empty order id",
			req: servicedto.PayOrderRequest{
				PaymentMethod: domain.PaymentMethodCard,
			},
			wantErr: domain.ErrOrderIDRequired,
		},
		{
			name: "error invalid payment method",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodUnknown,
			},
			wantErr: domain.ErrInvalidPaymentMethod,
		},
		{
			name: "error get order from repository",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			setupMocks: func(ctx context.Context, repository *mocks.Repository, _ *mocks.PaymentClient) {
				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(domain.Order{}, errRepository).
					Once()
			},
			wantErr: errRepository,
		},
		{
			name: "error order already paid",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			setupMocks: func(ctx context.Context, repository *mocks.Repository, _ *mocks.PaymentClient) {
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
			name: "error order cancelled",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			setupMocks: func(ctx context.Context, repository *mocks.Repository, _ *mocks.PaymentClient) {
				order := validOrder()
				order.Status = domain.OrderStatusCancelled

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()
			},
			wantErr: domain.ErrOrderCancelled,
		},
		{
			name: "error invalid order status",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			setupMocks: func(ctx context.Context, repository *mocks.Repository, _ *mocks.PaymentClient) {
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
			name: "error payment client",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			setupMocks: func(ctx context.Context, repository *mocks.Repository, paymentClient *mocks.PaymentClient) {
				order := validOrder()

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()

				paymentClient.EXPECT().
					PayOrder(ctx, clientdto.PayOrderRequest{
						OrderID:       orderID,
						UserID:        userID,
						PaymentMethod: domain.PaymentMethodCard,
					}).
					Return(clientdto.PayOrderResponse{}, errClient).
					Once()
			},
			wantErr: errClient,
		},
		{
			name: "error update order in repository",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			setupMocks: func(ctx context.Context, repository *mocks.Repository, paymentClient *mocks.PaymentClient) {
				order := validOrder()
				expectedOrder := paidOrder(domain.PaymentMethodCard)

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()

				paymentClient.EXPECT().
					PayOrder(ctx, clientdto.PayOrderRequest{
						OrderID:       orderID,
						UserID:        userID,
						PaymentMethod: domain.PaymentMethodCard,
					}).
					Return(clientdto.PayOrderResponse{TransactionID: transactionID}, nil).
					Once()

				repository.EXPECT().
					UpdateOrder(ctx, expectedOrder).
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
			paymentClient := mocks.NewPaymentClient(t)
			service := NewService(repository, nil, paymentClient)

			if tt.setupMocks != nil {
				tt.setupMocks(ctx, repository, paymentClient)
			}

			got, err := service.PayOrder(ctx, tt.req)

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

func TestService_PayOrder_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repository := mocks.NewRepository(t)
	paymentClient := mocks.NewPaymentClient(t)
	service := NewService(repository, nil, paymentClient)

	got, err := service.PayOrder(ctx, servicedto.PayOrderRequest{
		OrderID:       orderID,
		PaymentMethod: domain.PaymentMethodCard,
	})

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, domain.Order{}, got)
}
