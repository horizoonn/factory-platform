package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/client/dto"
	"github.com/horizoonn/factory-platform/order/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
)

func (s *Service) PayOrder(ctx context.Context, req servicedto.PayOrderRequest) (domain.Order, error) {
	if err := s.validateContext(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("pay order: %w", err)
	}

	if req.OrderID == uuid.Nil {
		return domain.Order{}, domain.ErrOrderIDRequired
	}

	if !req.PaymentMethod.Valid() {
		return domain.Order{}, domain.ErrInvalidPaymentMethod
	}

	order, err := s.repository.GetOrder(ctx, req.OrderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order with id='%s' from repository: %w", req.OrderID, err)
	}

	switch order.Status {
	case domain.OrderStatusPaid:
		return domain.Order{}, domain.ErrOrderAlreadyPaid
	case domain.OrderStatusCancelled:
		return domain.Order{}, domain.ErrOrderCancelled
	case domain.OrderStatusPendingPayment:
	default:
		return domain.Order{}, domain.ErrInvalidOrderStatus
	}

	payment, err := s.paymentClient.PayOrder(ctx, dto.PayOrderRequest{
		OrderID:       req.OrderID,
		UserID:        order.UserID,
		PaymentMethod: req.PaymentMethod,
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf("pay order through payment client: %w", err)
	}

	order.TransactionID = &payment.TransactionID
	order.PaymentMethod = &req.PaymentMethod
	order.Status = domain.OrderStatusPaid

	updatedOrder, err := s.repository.UpdateOrder(ctx, order)
	if err != nil {
		return domain.Order{}, fmt.Errorf("update paid order: %w", err)
	}

	return updatedOrder, nil
}
