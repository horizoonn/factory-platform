package order

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/outbox"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
	"github.com/horizoonn/factory-platform/order/internal/service/mocks"
)

func TestService_PayOrder(t *testing.T) {
	tests := []struct {
		name       string
		req        servicedto.PayOrderRequest
		setupMocks func(
			ctx context.Context,
			repository *mocks.Repository,
			paymentClient *mocks.PaymentClient,
			orderPaidEncoder *mocks.OrderPaidEncoder,
		)
		want    domain.Order
		wantErr error
	}{
		{
			name: "success",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			setupMocks: func(
				ctx context.Context,
				repository *mocks.Repository,
				paymentClient *mocks.PaymentClient,
				orderPaidEncoder *mocks.OrderPaidEncoder,
			) {
				order := validOrder()
				expectedOrder := paidOrder(domain.PaymentMethodCard)
				expectedEvent := domain.OrderPaidEvent{
					ID:            eventID,
					OrderID:       orderID,
					UserID:        userID,
					PaymentMethod: domain.PaymentMethodCard,
					TransactionID: transactionID,
					OccurredAt:    eventTime,
				}
				encodedEvent := outbox.Event{ID: eventID}

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()

				paymentClient.EXPECT().
					PayOrder(ctx, servicedto.PaymentRequest{
						OrderID:       orderID,
						UserID:        userID,
						PaymentMethod: domain.PaymentMethodCard,
					}).
					Return(servicedto.PaymentResponse{TransactionID: transactionID}, nil).
					Once()

				orderPaidEncoder.EXPECT().
					Encode(expectedEvent).
					Return(encodedEvent, nil).
					Once()

				repository.EXPECT().
					MarkPaidAndEnqueueOrderPaid(ctx, expectedOrder, encodedEvent).
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
			setupMocks: func(ctx context.Context, repository *mocks.Repository, _ *mocks.PaymentClient, _ *mocks.OrderPaidEncoder) {
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
			setupMocks: func(ctx context.Context, repository *mocks.Repository, _ *mocks.PaymentClient, _ *mocks.OrderPaidEncoder) {
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
			setupMocks: func(ctx context.Context, repository *mocks.Repository, _ *mocks.PaymentClient, _ *mocks.OrderPaidEncoder) {
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
			setupMocks: func(ctx context.Context, repository *mocks.Repository, _ *mocks.PaymentClient, _ *mocks.OrderPaidEncoder) {
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
			setupMocks: func(ctx context.Context, repository *mocks.Repository, paymentClient *mocks.PaymentClient, _ *mocks.OrderPaidEncoder) {
				order := validOrder()

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()

				paymentClient.EXPECT().
					PayOrder(ctx, servicedto.PaymentRequest{
						OrderID:       orderID,
						UserID:        userID,
						PaymentMethod: domain.PaymentMethodCard,
					}).
					Return(servicedto.PaymentResponse{}, errClient).
					Once()
			},
			wantErr: errClient,
		},
		{
			name: "error encode OrderPaid event",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			setupMocks: func(
				ctx context.Context,
				repository *mocks.Repository,
				paymentClient *mocks.PaymentClient,
				orderPaidEncoder *mocks.OrderPaidEncoder,
			) {
				order := validOrder()
				expectedEvent := domain.OrderPaidEvent{
					ID:            eventID,
					OrderID:       orderID,
					UserID:        userID,
					PaymentMethod: domain.PaymentMethodCard,
					TransactionID: transactionID,
					OccurredAt:    eventTime,
				}

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()

				paymentClient.EXPECT().
					PayOrder(ctx, servicedto.PaymentRequest{
						OrderID:       orderID,
						UserID:        userID,
						PaymentMethod: domain.PaymentMethodCard,
					}).
					Return(servicedto.PaymentResponse{TransactionID: transactionID}, nil).
					Once()

				orderPaidEncoder.EXPECT().
					Encode(expectedEvent).
					Return(outbox.Event{}, errEncoder).
					Once()
			},
			wantErr: errEncoder,
		},
		{
			name: "error mark paid and enqueue event",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			setupMocks: func(
				ctx context.Context,
				repository *mocks.Repository,
				paymentClient *mocks.PaymentClient,
				orderPaidEncoder *mocks.OrderPaidEncoder,
			) {
				order := validOrder()
				expectedOrder := paidOrder(domain.PaymentMethodCard)
				expectedEvent := domain.OrderPaidEvent{
					ID:            eventID,
					OrderID:       orderID,
					UserID:        userID,
					PaymentMethod: domain.PaymentMethodCard,
					TransactionID: transactionID,
					OccurredAt:    eventTime,
				}
				encodedEvent := outbox.Event{ID: eventID}

				repository.EXPECT().
					GetOrder(ctx, orderID).
					Return(order, nil).
					Once()

				paymentClient.EXPECT().
					PayOrder(ctx, servicedto.PaymentRequest{
						OrderID:       orderID,
						UserID:        userID,
						PaymentMethod: domain.PaymentMethodCard,
					}).
					Return(servicedto.PaymentResponse{TransactionID: transactionID}, nil).
					Once()

				orderPaidEncoder.EXPECT().
					Encode(expectedEvent).
					Return(encodedEvent, nil).
					Once()

				repository.EXPECT().
					MarkPaidAndEnqueueOrderPaid(ctx, expectedOrder, encodedEvent).
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
			orderPaidEncoder := mocks.NewOrderPaidEncoder(t)
			service := NewService(repository, nil, paymentClient, orderPaidEncoder)
			service.idGenerator = func() uuid.UUID { return eventID }
			service.clock = func() time.Time { return eventTime }

			if tt.setupMocks != nil {
				tt.setupMocks(ctx, repository, paymentClient, orderPaidEncoder)
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
	orderPaidEncoder := mocks.NewOrderPaidEncoder(t)
	service := NewService(repository, nil, paymentClient, orderPaidEncoder)

	got, err := service.PayOrder(ctx, servicedto.PayOrderRequest{
		OrderID:       orderID,
		PaymentMethod: domain.PaymentMethodCard,
	})

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, domain.Order{}, got)
}
