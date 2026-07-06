package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/client/dto"
	"github.com/horizoonn/factory-platform/order/internal/domain"
)

type OrderService struct {
	orderRepository OrderRepository
	inventoryClient InventoryClient
	paymentClient   PaymentClient
	idGenerator     IDGenerator
}

func NewOrderService(repo OrderRepository, inventoryClient InventoryClient, paymentClient PaymentClient) *OrderService {
	return &OrderService{
		orderRepository: repo,
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
		idGenerator:     uuid.New,
	}
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (domain.Order, error)
	UpdateOrder(ctx context.Context, order domain.Order) (domain.Order, error)
}

type InventoryClient interface {
	ListParts(ctx context.Context, partIDs []uuid.UUID) ([]domain.Part, error)
}

type PaymentClient interface {
	PayOrder(ctx context.Context, req dto.PayOrderRequest) (dto.PayOrderResponse, error)
}

type IDGenerator func() uuid.UUID

type CreateOrderRequest struct {
	UserID  uuid.UUID
	PartIDs []uuid.UUID
}
