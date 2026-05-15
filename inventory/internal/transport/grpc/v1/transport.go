package grpcv1

import (
	"context"

	"github.com/google/uuid"
	"github.com/horizoonn/factory-platform.git/inventory/internal/domain"
	inventoryv1 "github.com/horizoonn/factory-platform.git/shared/pkg/proto/inventory/v1"
)

type InventoryServer struct {
	inventoryv1.UnimplementedInventoryServiceServer

	inventoryService InventoryService
}

type InventoryService interface {
	GetPart(ctx context.Context, id uuid.UUID) (domain.Part, error)
	ListParts(ctx context.Context, filter domain.PartsFilter) ([]domain.Part, error)
}

func NewInventoryServer(inventoryService InventoryService) *InventoryServer {
	return &InventoryServer{
		inventoryService: inventoryService,
	}
}
