package model

import (
	"time"

	"github.com/google/uuid"

	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

type Part struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Price         float64
	StockQuantity int64
	Category      int32
	Dimensions    *Dimensions
	Manufacturer  *Manufacturer
	Tags          []string
	Metadata      map[string]*Value
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

type Dimensions struct {
	Length float64
	Width  float64
	Height float64
	Weight float64
}

type Manufacturer struct {
	Name    string
	Country string
	Website string
}

type Value struct {
	String  *string
	Int64   *int64
	Float64 *float64
	Bool    *bool
}

func (m *Part) Scan(row postgrespool.Row) error {
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
