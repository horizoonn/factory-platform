package order

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/outbox"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
)

type Service struct {
	repository       Repository
	inventoryClient  InventoryClient
	paymentClient    PaymentClient
	orderPaidEncoder OrderPaidEncoder
	idGenerator      IDGenerator
	clock            Clock
}

func NewService(
	repository Repository,
	inventoryClient InventoryClient,
	paymentClient PaymentClient,
	orderPaidEncoder OrderPaidEncoder,
) *Service {
	return &Service{
		repository:       repository,
		inventoryClient:  inventoryClient,
		paymentClient:    paymentClient,
		orderPaidEncoder: orderPaidEncoder,
		idGenerator:      uuid.New,
		clock:            time.Now,
	}
}

type Repository interface {
	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error)
	GetOrder(ctx context.Context, orderID uuid.UUID) (domain.Order, error)
	CancelOrder(ctx context.Context, orderID uuid.UUID) error
	MarkPaidAndEnqueueOrderPaid(
		ctx context.Context,
		order domain.Order,
		event outbox.Event,
	) (domain.Order, error)
	CompleteOrder(ctx context.Context, event domain.ShipAssembledEvent) error
}

type InventoryClient interface {
	ListParts(ctx context.Context, partIDs []uuid.UUID) ([]domain.Part, error)
}

type PaymentClient interface {
	PayOrder(ctx context.Context, req servicedto.PaymentRequest) (servicedto.PaymentResponse, error)
}

type OrderPaidEncoder interface {
	Encode(event domain.OrderPaidEvent) (outbox.Event, error)
}

type IDGenerator func() uuid.UUID

type Clock func() time.Time

func (s *Service) validateContext(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
