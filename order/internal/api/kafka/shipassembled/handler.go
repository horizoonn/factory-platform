package shipassembled

import (
	"context"
	"errors"
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer"
)

type OrderService interface {
	CompleteOrder(ctx context.Context, event domain.ShipAssembledEvent) error
}

type Handler struct {
	service OrderService
}

func NewHandler(service OrderService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Handle(ctx context.Context, record kafka.Record) error {
	event, err := recordToShipAssembledEvent(record)
	if err != nil {
		return consumer.Permanent(fmt.Errorf("decode ShipAssembled record: %w", err))
	}

	if err := h.service.CompleteOrder(ctx, event); err != nil {
		wrappedErr := fmt.Errorf("complete assembled order: %w", err)
		if isPermanentCompleteOrderError(err) {
			return consumer.Permanent(wrappedErr)
		}

		return wrappedErr
	}

	return nil
}

func isPermanentCompleteOrderError(err error) bool {
	return errors.Is(err, domain.ErrEventIDRequired) ||
		errors.Is(err, domain.ErrOrderIDRequired) ||
		errors.Is(err, domain.ErrUserIDRequired) ||
		errors.Is(err, domain.ErrOccurredAtRequired) ||
		errors.Is(err, domain.ErrInvalidBuildTime) ||
		errors.Is(err, domain.ErrNotFound) ||
		errors.Is(err, domain.ErrOrderUserMismatch) ||
		errors.Is(err, domain.ErrInvalidOrderStatus)
}
