package inventoryv1

import (
	"context"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

type Server struct {
	inventorypb.UnimplementedInventoryServiceServer

	service InventoryService
}

type InventoryService interface {
	GetPart(ctx context.Context, id uuid.UUID) (domain.Part, error)
	ListParts(ctx context.Context, filter domain.PartsFilter) ([]domain.Part, error)
}

func NewServer(service InventoryService) *Server {
	return &Server{
		service: service,
	}
}
