package payment

import (
	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/payment/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/payment/internal/service/dto"
)

var (
	orderID       = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID        = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	transactionID = uuid.MustParse("00000000-0000-0000-0000-000000000003")
)

func validPayOrderRequest() servicedto.PayOrderRequest {
	return servicedto.PayOrderRequest{
		OrderID:       orderID,
		UserID:        userID,
		PaymentMethod: domain.PaymentMethodCard,
	}
}

func newServiceWithTransactionID(transactionID uuid.UUID) *Service {
	service := NewService()
	service.transactionIDGenerator = func() uuid.UUID {
		return transactionID
	}
	return service
}
