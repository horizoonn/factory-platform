package dto

import (
	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
)

type PayOrderRequest struct {
	OrderID       uuid.UUID
	PaymentMethod domain.PaymentMethod
}

type PaymentRequest struct {
	OrderID       uuid.UUID
	UserID        uuid.UUID
	PaymentMethod domain.PaymentMethod
}

type PaymentResponse struct {
	TransactionID uuid.UUID
}
