package inventoryv1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	inventorypb "github.com/horizoonn/factory-platform.git/shared/pkg/proto/inventory/v1"
)

func (s *InventoryServer) ListParts(ctx context.Context, req *inventorypb.ListPartsRequest) (*inventorypb.ListPartsResponse, error) {
	filter, err := filterToDomain(req.GetFilter())
	if err != nil {
		return nil, err
	}
	parts, err := s.inventoryService.ListParts(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &inventorypb.ListPartsResponse{
		Parts: partsToProto(parts),
	}, nil
}
