package orderpaid

import (
	"context"
	"fmt"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer"
)

type AssemblyService interface {
	Assemble(ctx context.Context, event domain.OrderPaidEvent) error
}

type Handler struct {
	service AssemblyService
}

func NewHandler(service AssemblyService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Handle(ctx context.Context, record kafka.Record) error {
	event, err := recordToOrderPaidEvent(record)
	if err != nil {
		return consumer.Permanent(fmt.Errorf("decode OrderPaid record: %w", err))
	}

	if err := h.service.Assemble(ctx, event); err != nil {
		return fmt.Errorf("assemble paid order: %w", err)
	}

	return nil
}
