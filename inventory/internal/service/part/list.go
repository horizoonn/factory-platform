package service

import (
	"context"
	"fmt"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
)

func (s *Service) ListParts(ctx context.Context, filter domain.PartsFilter) ([]domain.Part, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("list parts context: %w", err)
	}

	parts, err := s.repository.ListParts(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list parts from repository: %w", err)
	}

	return parts, nil
}
