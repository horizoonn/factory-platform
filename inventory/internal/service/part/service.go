package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
)

type Service struct {
	repository Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repository: repo,
	}
}

type Repository interface {
	GetPart(ctx context.Context, id uuid.UUID) (domain.Part, error)
	ListParts(ctx context.Context, filter domain.PartsFilter) ([]domain.Part, error)
}
