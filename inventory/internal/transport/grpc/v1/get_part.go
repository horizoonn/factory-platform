package grpcv1

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/horizoonn/factory-platform.git/inventory/internal/domain"
	inventoryv1 "github.com/horizoonn/factory-platform.git/shared/pkg/proto/inventory/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InventoryServer) GetPart(ctx context.Context, req *inventoryv1.GetPartRequest) (*inventoryv1.GetPartResponse, error) {
	id, err := uuid.Parse(req.GetUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid")
	}

	part, err := s.inventoryService.GetPart(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrPartNotFound) {
			return nil, status.Errorf(codes.NotFound, "part with id='%s' not found", req.GetUuid())
		}

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &inventoryv1.GetPartResponse{
		Part: partToProto(&part),
	}, nil
}
