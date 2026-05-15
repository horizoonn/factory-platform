package grpcv1

import (
	"context"

	inventoryv1 "github.com/horizoonn/factory-platform.git/shared/pkg/proto/inventory/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InventoryServer) ListParts(ctx context.Context, req *inventoryv1.ListPartsRequest) (*inventoryv1.ListPartsResponse, error) {
	filter, err := filterToDomain(req.GetFilter())
	if err != nil {
		return nil, err
	}
	parts, err := s.inventoryService.ListParts(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &inventoryv1.ListPartsResponse{
		Parts: partsToProto(parts),
	}, nil
}
