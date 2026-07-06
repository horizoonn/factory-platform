package domain

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	PartIDs       []uuid.UUID
	TotalPrice    float64
	TransactionID *uuid.UUID
	PaymentMethod *PaymentMethod
	Status        OrderStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
