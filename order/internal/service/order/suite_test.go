package order

import (
	"errors"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
)

var (
	orderID       = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID        = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	partID        = uuid.MustParse("00000000-0000-0000-0000-000000000003")
	transactionID = uuid.MustParse("00000000-0000-0000-0000-000000000004")

	errRepository = errors.New("repository error")
	errClient     = errors.New("client error")
)

func validCreateOrderRequest() servicedto.CreateOrderRequest {
	return servicedto.CreateOrderRequest{
		UserID:  userID,
		PartIDs: []uuid.UUID{partID},
	}
}

func validOrder() domain.Order {
	return domain.Order{
		ID:         orderID,
		UserID:     userID,
		PartIDs:    []uuid.UUID{partID},
		TotalPrice: 100,
		Status:     domain.OrderStatusPendingPayment,
	}
}

func paidOrder(paymentMethod domain.PaymentMethod) domain.Order {
	order := validOrder()
	order.TransactionID = &transactionID
	order.PaymentMethod = &paymentMethod
	order.Status = domain.OrderStatusPaid
	return order
}

func validPart() domain.Part {
	return domain.Part{
		ID:    partID,
		Price: 100,
	}
}

func newServiceWithOrderID(
	repository Repository,
	inventoryClient InventoryClient,
	paymentClient PaymentClient,
	orderID uuid.UUID,
) *Service {
	service := NewService(repository, inventoryClient, paymentClient)
	service.idGenerator = func() uuid.UUID {
		return orderID
	}
	return service
}

func newServiceWithRepository(repository Repository) *Service {
	return NewService(repository, nil, nil)
}
