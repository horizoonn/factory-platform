package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/payment/internal/domain"
)

func (s *PaymentService) PayOrder(ctx context.Context, req PayOrderRequest) (PayOrderResponse, error) {
	if err := ctx.Err(); err != nil {
		return PayOrderResponse{}, fmt.Errorf("pay order context: %w", err)
	}

	if req.OrderID == uuid.Nil {
		return PayOrderResponse{}, domain.ErrOrderIDRequired
	}

	if req.UserID == uuid.Nil {
		return PayOrderResponse{}, domain.ErrUserIDRequired
	}

	if !req.PaymentMethod.Valid() {
		return PayOrderResponse{}, domain.ErrInvalidPaymentMethod
	}

	transactionIDGenerator := s.transactionIDGenerator
	if transactionIDGenerator == nil {
		transactionIDGenerator = uuid.New
	}

	return PayOrderResponse{
		TransactionID: transactionIDGenerator(),
	}, nil
}
