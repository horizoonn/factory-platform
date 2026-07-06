package inventoryv1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/horizoonn/factory-platform/inventory/internal/converter"
	"github.com/horizoonn/factory-platform/inventory/internal/domain"
	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

func (s *Server) ListParts(ctx context.Context, req *inventorypb.ListPartsRequest) (*inventorypb.ListPartsResponse, error) {
	filter, err := converter.FilterToDomain(req.GetFilter())
	if err != nil {
		return nil, err
	}

	parts, err := s.service.ListParts(ctx, filter)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "parts not found")
		}
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &inventorypb.ListPartsResponse{
		Parts: converter.PartsToProto(parts),
	}, nil
}
