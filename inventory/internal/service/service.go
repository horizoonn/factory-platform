package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/horizoonn/factory-platform.git/inventory/internal/domain"
)

type InventoryService struct {
	inventoryRepository InventoryRepository
}

func NewInventoryService(repo InventoryRepository) *InventoryService {
	return &InventoryService{
		inventoryRepository: repo,
	}
}

type InventoryRepository interface {
	GetPart(ctx context.Context, id uuid.UUID) (domain.Part, error)
	ListParts(ctx context.Context, filter domain.PartsFilter) ([]domain.Part, error)
}
