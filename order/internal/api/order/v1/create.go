package orderv1

import (
	"context"
	"errors"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

func (h *Handler) CreateOrder(
	ctx context.Context,
	req *orderopenapi.CreateOrderRequest,
) (orderopenapi.CreateOrderRes, error) {
	if h.service == nil {
		return nil, domain.ErrNotImplemented
	}

	order, err := h.service.CreateOrder(ctx, servicedto.CreateOrderRequest{
		UserID:  req.UserUUID,
		PartIDs: req.PartUuids,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserIDRequired) {
			return badRequest("user id is required"), nil
		}
		if errors.Is(err, domain.ErrEmptyParts) {
			return badRequest("parts list is empty"), nil
		}
		if errors.Is(err, domain.ErrPartsNotFound) {
			return badRequest("some parts not found"), nil
		}
		return nil, err
	}

	return &orderopenapi.CreateOrderResponse{
		OrderUUID:  order.ID,
		TotalPrice: order.TotalPrice,
	}, nil
}
