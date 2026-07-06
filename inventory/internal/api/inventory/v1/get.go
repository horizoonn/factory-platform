package inventoryv1

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

func (s *InventoryServer) GetPart(ctx context.Context, req *inventorypb.GetPartRequest) (*inventorypb.GetPartResponse, error) {
	id, err := uuid.Parse(req.GetUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid")
	}

	part, err := s.inventoryService.GetPart(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "part with id='%s' not found", req.GetUuid())
		}

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &inventorypb.GetPartResponse{
		Part: partToProto(&part),
	}, nil
}
