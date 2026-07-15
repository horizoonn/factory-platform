package postgrespool

import (
	"context"
	"time"
)

type Pool interface {
	Executor

	Ping(ctx context.Context) error
	Close()

	OpTimeout() time.Duration
}

type Executor interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, arguments ...any) (CommandTag, error)
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(Executor) error) error
}

type TransactionalPool interface {
	Pool
	Transactor
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type Row interface {
	Scan(dest ...any) error
}

type CommandTag interface {
	RowsAffected() int64
}
