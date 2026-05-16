package part

import (
	"time"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform.git/inventory/internal/domain"
	"github.com/horizoonn/factory-platform.git/platform/pkg/database/postgres/pool"
)

type partModel struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Price         float64
	StockQuantity int64
	Category      int32
	Dimensions    *domain.Dimensions
	Manufacturer  *domain.Manufacturer
	Tags          []string
	Metadata      map[string]*domain.Value
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

func (m *partModel) scan(row postgrespool.Row) error {
	return row.Scan(
		&m.ID,
		&m.Name,
		&m.Description,
		&m.Price,
		&m.StockQuantity,
		&m.Category,
		&m.Dimensions,
		&m.Manufacturer,
		&m.Tags,
		&m.Metadata,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}

func partModelToDomain(part partModel) domain.Part {
	domainPart := domain.NewPart(
		part.ID,
		part.Name,
		part.Description,
		part.Price,
		part.StockQuantity,
		part.Category,
		part.Dimensions,
		part.Manufacturer,
		part.Tags,
		part.Metadata,
		part.CreatedAt,
		part.UpdatedAt,
	)

	return *domainPart
}
