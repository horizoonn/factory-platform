package payment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/payment/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/payment/internal/service/dto"
)

func TestService_PayOrder(t *testing.T) {
	tests := []struct {
		name    string
		req     servicedto.PayOrderRequest
		want    servicedto.PayOrderResponse
		wantErr error
	}{
		{
			name: "success card",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				UserID:        userID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			want: servicedto.PayOrderResponse{
				TransactionID: transactionID,
			},
		},
		{
			name: "success sbp",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				UserID:        userID,
				PaymentMethod: domain.PaymentMethodSBP,
			},
			want: servicedto.PayOrderResponse{
				TransactionID: transactionID,
			},
		},
		{
			name: "success credit card",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				UserID:        userID,
				PaymentMethod: domain.PaymentMethodCreditCard,
			},
			want: servicedto.PayOrderResponse{
				TransactionID: transactionID,
			},
		},
		{
			name: "success investor money",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				UserID:        userID,
				PaymentMethod: domain.PaymentMethodInvestorMoney,
			},
			want: servicedto.PayOrderResponse{
				TransactionID: transactionID,
			},
		},
		{
			name: "error empty user id",
			req: servicedto.PayOrderRequest{
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			wantErr: domain.ErrUserIDRequired,
		},
		{
			name: "error empty order id",
			req: servicedto.PayOrderRequest{
				UserID:        userID,
				PaymentMethod: domain.PaymentMethodCard,
			},
			wantErr: domain.ErrOrderIDRequired,
		},
		{
			name: "error invalid payment method",
			req: servicedto.PayOrderRequest{
				UserID:        userID,
				OrderID:       orderID,
				PaymentMethod: domain.PaymentMethodUnknown,
			},
			wantErr: domain.ErrInvalidPaymentMethod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newServiceWithTransactionID(transactionID)

			got, err := service.PayOrder(context.Background(), tt.req)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, servicedto.PayOrderResponse{}, got)
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

	service := newServiceWithTransactionID(transactionID)

	got, err := service.PayOrder(ctx, validPayOrderRequest())

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, servicedto.PayOrderResponse{}, got)
}
