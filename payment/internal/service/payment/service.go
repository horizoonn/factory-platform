package payment

import (
	"github.com/google/uuid"
)

type Service struct {
	transactionIDGenerator TransactionIDGenerator
}

func NewService() *Service {
	return &Service{
		transactionIDGenerator: uuid.New,
	}
}

type TransactionIDGenerator func() uuid.UUID
