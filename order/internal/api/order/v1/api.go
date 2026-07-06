package orderv1

import (
	"context"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
)

type Handler struct {
	service OrderService
}

type OrderService interface {
	CreateOrder(ctx context.Context, req servicedto.CreateOrderRequest) (domain.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (domain.Order, error)
	PayOrder(ctx context.Context, req servicedto.PayOrderRequest) (domain.Order, error)
	CancelOrder(ctx context.Context, id uuid.UUID) error
}

func NewHandler(service OrderService) *Handler {
	return &Handler{
		service: service,
	}
}
