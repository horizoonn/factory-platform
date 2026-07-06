package domain

import "github.com/google/uuid"

type Part struct {
	ID    uuid.UUID
	Price float64
}
