package service

import (
	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/payment/internal/domain"
)

type PaymentService struct {
	transactionIDGenerator TransactionIDGenerator
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		transactionIDGenerator: uuid.New,
	}
}

type PayOrderRequest struct {
	OrderID       uuid.UUID
	UserID        uuid.UUID
	PaymentMethod domain.PaymentMethod
}

type PayOrderResponse struct {
	TransactionID uuid.UUID
}

type TransactionIDGenerator func() uuid.UUID
