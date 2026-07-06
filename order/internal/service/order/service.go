package order

import (
	"context"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/client/dto"
	"github.com/horizoonn/factory-platform/order/internal/domain"
)

type Service struct {
	repository      Repository
	inventoryClient InventoryClient
	paymentClient   PaymentClient
	idGenerator     IDGenerator
}

func NewService(repo Repository, inventoryClient InventoryClient, paymentClient PaymentClient) *Service {
	return &Service{
		repository:      repo,
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
		idGenerator:     uuid.New,
	}
}

type Repository interface {
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

func (s *Service) validateContext(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
