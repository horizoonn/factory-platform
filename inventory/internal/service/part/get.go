package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
)

func (s *Service) GetPart(ctx context.Context, id uuid.UUID) (domain.Part, error) {
	if err := ctx.Err(); err != nil {
		return domain.Part{}, fmt.Errorf("get part context: %w", err)
	}

	if id == uuid.Nil {
		return domain.Part{}, domain.ErrPartIDRequired
	}

	part, err := s.repository.GetPart(ctx, id)
	if err != nil {
		return domain.Part{}, fmt.Errorf("get part with id='%s' from repository: %w", id, err)
	}

	return part, nil
}
