package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform.git/inventory/internal/domain"
)

func (s *InventoryService) GetPart(ctx context.Context, id uuid.UUID) (domain.Part, error) {
	part, err := s.inventoryRepository.GetPart(ctx, id)
	if err != nil {
		return domain.Part{}, fmt.Errorf("get part with id='%s' from repository: %w", id, err)
	}

	return part, nil
}
