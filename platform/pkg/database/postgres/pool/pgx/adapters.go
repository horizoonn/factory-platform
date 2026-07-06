package pgxpool

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	postgresPool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

type pgxRows struct {
	pgx.Rows
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

	const pgxForeignKeyErrorCode = "23503"

	if errors.Is(err, pgx.ErrNoRows) {
		return postgresPool.ErrNoRows
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgxForeignKeyErrorCode {
			return fmt.Errorf("%w: %w", postgresPool.ErrViolatesForeignKey, err)
		}
	}

	return err
}
