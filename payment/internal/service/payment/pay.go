package payment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/payment/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/payment/internal/service/dto"
)

func (s *Service) PayOrder(ctx context.Context, req servicedto.PayOrderRequest) (servicedto.PayOrderResponse, error) {
	if err := s.validateContext(ctx); err != nil {
		return servicedto.PayOrderResponse{}, fmt.Errorf("pay order: %w", err)
	}

	if req.OrderID == uuid.Nil {
		return servicedto.PayOrderResponse{}, domain.ErrOrderIDRequired
	}

	if req.UserID == uuid.Nil {
		return servicedto.PayOrderResponse{}, domain.ErrUserIDRequired
	}

	if !req.PaymentMethod.Valid() {
		return servicedto.PayOrderResponse{}, domain.ErrInvalidPaymentMethod
	}

	idGenerator := s.transactionIDGenerator
	if idGenerator == nil {
		idGenerator = uuid.New
	}

	return servicedto.PayOrderResponse{
		TransactionID: idGenerator(),
	}, nil
}

func (s *Service) validateContext(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
