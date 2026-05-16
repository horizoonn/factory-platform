package pgxpool

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	postgresPool "github.com/horizoonn/factory-platform.git/platform/pkg/database/postgres/pool"
)

type pgxRows struct {
	pgx.Rows
}

func (r pgxRows) Scan(dest ...any) error {
	return mapErrors(r.Rows.Scan(dest...))
}

func (r pgxRows) Err() error {
	return mapErrors(r.Rows.Err())
}

type pgxRow struct {
	pgx.Row
}

func (r pgxRow) Scan(dest ...any) error {
	return mapErrors(r.Row.Scan(dest...))
}

type pgxCommandTag struct {
	pgconn.CommandTag
}

func mapErrors(err error) error {
	if err == nil {
		return nil
	}

	const (
		pgxViolatesForeignKeyErrorCode = "23503"
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return postgresPool.ErrNoRows
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgxViolatesForeignKeyErrorCode {
			return fmt.Errorf("%w: %w", postgresPool.ErrViolatesForeignKey, err)
		}
	}

	return fmt.Errorf("%w: %w", postgresPool.ErrUnknown, err)
}
