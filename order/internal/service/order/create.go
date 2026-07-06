package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
)

func (s *Service) CreateOrder(ctx context.Context, req servicedto.CreateOrderRequest) (domain.Order, error) {
	if err := s.validateContext(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("create order: %w", err)
	}

	if req.UserID == uuid.Nil {
		return domain.Order{}, domain.ErrUserIDRequired
	}

	if len(req.PartIDs) == 0 {
		return domain.Order{}, domain.ErrEmptyParts
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

	createdOrder, err := s.repository.CreateOrder(ctx, order)
	if err != nil {
		return domain.Order{}, fmt.Errorf("create order in repository: %w", err)
	}

	return createdOrder, nil
}
