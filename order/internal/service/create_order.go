package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
)

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (domain.Order, error) {
	if err := ctx.Err(); err != nil {
		return domain.Order{}, fmt.Errorf("create order context: %w", err)
	}

	if req.UserID == uuid.Nil {
		return domain.Order{}, domain.ErrUserIDRequired
	}

	if len(req.PartIDs) == 0 {
		return domain.Order{}, domain.ErrEmptyParts
	}

	if s.inventoryClient == nil || s.orderRepository == nil {
		return domain.Order{}, domain.ErrNotImplemented
	}

	parts, err := s.inventoryClient.ListParts(ctx, req.PartIDs)
	if err != nil {
		return domain.Order{}, fmt.Errorf("list order parts: %w", err)
	}

	foundParts := make(map[uuid.UUID]domain.Part, len(parts))
	for _, part := range parts {
		foundParts[part.ID] = part
	}

	totalPrice := 0.0
	for _, requestedID := range req.PartIDs {
		part, ok := foundParts[requestedID]
		if !ok {
			return domain.Order{}, domain.ErrPartsNotFound
		}

		totalPrice += part.Price
	}

	idGenerator := s.idGenerator
	if idGenerator == nil {
		idGenerator = uuid.New
	}

	order := domain.Order{
		ID:         idGenerator(),
		UserID:     req.UserID,
		PartIDs:    req.PartIDs,
		TotalPrice: totalPrice,
		Status:     domain.OrderStatusPendingPayment,
	}

	createdOrder, err := s.orderRepository.CreateOrder(ctx, order)
	if err != nil {
		return domain.Order{}, fmt.Errorf("create order in repository: %w", err)
	}

	return createdOrder, nil
}
