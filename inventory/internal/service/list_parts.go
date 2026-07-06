package service

import (
	"context"
	"fmt"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
)

func (s *InventoryService) ListParts(ctx context.Context, filter domain.PartsFilter) ([]domain.Part, error) {
	parts, err := s.inventoryRepository.ListParts(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get parts from repository: %w", err)
	}

	return parts, nil
}
