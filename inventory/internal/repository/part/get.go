package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform.git/inventory/internal/domain"
	postgrespool "github.com/horizoonn/factory-platform.git/platform/pkg/database/postgres/pool"
)

func (r *Repository) GetPart(ctx context.Context, id uuid.UUID) (domain.Part, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, name, description, price, stock_quantity, category, dimensions,
		manufacturer, tags, metadata, created_at, updated_at
	FROM platform.parts
	WHERE id=$1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var model partModel
	if err := model.scan(row); err != nil {
		if errors.Is(err, postgrespool.ErrNoRows) {
			return domain.Part{}, fmt.Errorf("part with id='%s': %w", id, domain.ErrNotFound)
		}

		return domain.Part{}, fmt.Errorf("scan part row: %w", err)
	}

	return partModelToDomain(model), nil
}
