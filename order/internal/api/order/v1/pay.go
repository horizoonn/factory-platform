package orderv1

import (
	"context"
	"errors"

	"github.com/horizoonn/factory-platform/order/internal/converter"
	"github.com/horizoonn/factory-platform/order/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

func (h *Handler) PayOrder(ctx context.Context, req *orderopenapi.PayOrderRequest, params orderopenapi.PayOrderParams) (orderopenapi.PayOrderRes, error) {
	if h.service == nil {
		return nil, domain.ErrNotImplemented
	}

	paymentMethod, err := converter.PaymentMethodToDomain(req.PaymentMethod)
	if err != nil {
		return badRequest("invalid payment method"), nil
	}

	order, err := h.service.PayOrder(ctx, servicedto.PayOrderRequest{
		OrderID:       params.OrderUUID,
		PaymentMethod: paymentMethod,
	})
	if err != nil {
		if errors.Is(err, domain.ErrOrderIDRequired) {
			return badRequest("order id is required"), nil
		}
		if errors.Is(err, domain.ErrInvalidPaymentMethod) {
			return badRequest("invalid payment method"), nil
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFound("order not found"), nil
		}
		if errors.Is(err, domain.ErrOrderAlreadyPaid) {
			return conflict("order already paid"), nil
		}
		if errors.Is(err, domain.ErrOrderCancelled) {
			return conflict("order is cancelled"), nil
		}
		if errors.Is(err, domain.ErrInvalidOrderStatus) {
			return conflict("order cannot be paid in current status"), nil
		}
		return nil, err
	}

	if order.TransactionID == nil {
		return internalError(), nil
	}

	return &orderopenapi.PayOrderResponse{
		TransactionUUID: *order.TransactionID,
	}, nil
}
