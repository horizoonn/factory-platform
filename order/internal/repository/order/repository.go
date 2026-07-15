package repository

import (
	"context"

	"github.com/horizoonn/factory-platform/order/internal/outbox"
	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

type Repository struct {
	pool   postgrespool.TransactionalPool
	outbox Outbox
}

func NewRepository(pool postgrespool.TransactionalPool, outbox Outbox) *Repository {
	return &Repository{
		pool:   pool,
		outbox: outbox,
	}
}

type Outbox interface {
	Enqueue(
		ctx context.Context,
		executor postgrespool.Executor,
		event outbox.Event,
	) (bool, error)
}
