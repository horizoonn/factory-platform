package orderv1

import (
	"context"
	"errors"
	"log/slog"

	"github.com/horizoonn/factory-platform/order/internal/converter"
	"github.com/horizoonn/factory-platform/order/internal/domain"
	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

func (h *Handler) GetOrder(
	ctx context.Context,
	params orderopenapi.GetOrderParams,
) (orderopenapi.GetOrderRes, error) {
	if h.service == nil {
		return nil, domain.ErrNotImplemented
	}

	order, err := h.service.GetOrder(ctx, params.OrderUUID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderIDRequired) {
			return badRequest("order id is required"), nil
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFound("order not found"), nil
		}
		return nil, err
	}

	dto, err := converter.OrderToOpenAPI(order)
	if err != nil {
		slog.Error(
			"corrupted order data in database",
			"order_id", order.ID,
			"status", order.Status,
			"error", err,
		)
		return internalError(), nil
	}

	return dto, nil
}
