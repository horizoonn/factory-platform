package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderPaidEvent struct {
	ID            uuid.UUID
	OrderID       uuid.UUID
	UserID        uuid.UUID
	PaymentMethod PaymentMethod
	TransactionID uuid.UUID
	OccurredAt    time.Time
}

type ShipAssembledEvent struct {
	ID           uuid.UUID
	OrderID      uuid.UUID
	UserID       uuid.UUID
	BuildTimeSec int64
	OccurredAt   time.Time
}
