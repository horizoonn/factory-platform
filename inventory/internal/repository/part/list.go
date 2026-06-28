package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/horizoonn/factory-platform.git/inventory/internal/domain"
)

func (r *Repository) ListParts(ctx context.Context, filter domain.PartsFilter) ([]domain.Part, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, name, description, price, stock_quantity, category, dimensions,
		manufacturer, tags, metadata, created_at, updated_at
	FROM platform.parts
	`

	conditions := make([]string, 0, 5)
	args := make([]any, 0, 5)
	addCondition := func(condition string, arg any) {
		args = append(args, arg)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	if len(filter.UUIDs) > 0 {
		addCondition("id = ANY($%d)", filter.UUIDs)
	}
	if len(filter.Names) > 0 {
		addCondition("name = ANY($%d)", filter.Names)
	}
	if len(filter.Categories) > 0 {
		categories := make([]int32, 0, len(filter.Categories))
		for _, category := range filter.Categories {
			categories = append(categories, int32(category))
		}
		addCondition("category = ANY($%d)", categories)
	}
	if len(filter.ManufacturerCountries) > 0 {
		addCondition("manufacturer->>'country' = ANY($%d)", filter.ManufacturerCountries)
	}
	if len(filter.Tags) > 0 {
		addCondition("tags && $%d", filter.Tags)
	}

	if len(conditions) > 0 {
		query += "WHERE " + strings.Join(conditions, "\n\tAND ")
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query parts: %w", err)
	}
	defer rows.Close()

	parts := make([]domain.Part, 0)
	for rows.Next() {
		var model partModel
		if err := model.scan(rows); err != nil {
			return nil, fmt.Errorf("scan part row: %w", err)
		}

		parts = append(parts, partModelToDomain(model))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate part rows: %w", err)
	}

	return parts, nil
}
