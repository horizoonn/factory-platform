package pgxpool

import (
	"context"

	"github.com/jackc/pgx/v5"

	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

type transactionExecutor struct {
	tx pgx.Tx
}

func (t *transactionExecutor) Query(ctx context.Context, sql string, args ...any) (postgrespool.Rows, error) {
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, mapErrors(err)
	}

	return pgxRows{rows}, nil
}

func (t *transactionExecutor) QueryRow(ctx context.Context, sql string, args ...any) postgrespool.Row {
	row := t.tx.QueryRow(ctx, sql, args...)

	return pgxRow{row}
}

func (t *transactionExecutor) Exec(ctx context.Context, sql string, arguments ...any) (postgrespool.CommandTag, error) {
	tag, err := t.tx.Exec(ctx, sql, arguments...)
	if err != nil {
		return nil, mapErrors(err)
	}

	return pgxCommandTag{tag}, nil
}

func (p *Pool) WithinTransaction(ctx context.Context, fn func(postgrespool.Executor) error) error {
	err := pgx.BeginFunc(ctx, p.Pool, func(tx pgx.Tx) error {
		executor := transactionExecutor{tx: tx}

		return fn(&executor)
	})

	return mapErrors(err)
}
