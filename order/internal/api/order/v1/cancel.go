package orderv1

import (
	"context"
	"errors"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

func (h *Handler) CancelOrder(ctx context.Context, params orderopenapi.CancelOrderParams) (orderopenapi.CancelOrderRes, error) {
	if h.service == nil {
		return nil, domain.ErrNotImplemented
	}

	if err := h.service.CancelOrder(ctx, params.OrderUUID); err != nil {
		if errors.Is(err, domain.ErrOrderIDRequired) {
			return badRequest("order id is required"), nil
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFound("order not found"), nil
		}
		if errors.Is(err, domain.ErrOrderAlreadyPaid) {
			return conflict("order already paid and cannot be canceled"), nil
		}
		if errors.Is(err, domain.ErrInvalidOrderStatus) {
			return conflict("order cannot be canceled in current status"), nil
		}
		return nil, err
	}

	return &orderopenapi.CancelOrderNoContent{}, nil
}
