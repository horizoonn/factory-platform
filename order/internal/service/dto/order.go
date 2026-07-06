package dto

import "github.com/google/uuid"

type CreateOrderRequest struct {
	UserID  uuid.UUID
	PartIDs []uuid.UUID
}
