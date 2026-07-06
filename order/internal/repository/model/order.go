package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

type Order struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	PartIDs       []uuid.UUID
	TotalPrice    float64
	TransactionID *uuid.UUID
	PaymentMethod sql.NullString
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (m *Order) Scan(row postgrespool.Row) error {
	return row.Scan(
		&m.ID,
		&m.UserID,
		&m.PartIDs,
		&m.TotalPrice,
		&m.TransactionID,
		&m.PaymentMethod,
		&m.Status,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}
