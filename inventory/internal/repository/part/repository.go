package repository

import (
	"github.com/horizoonn/factory-platform.git/platform/pkg/database/postgres/pool"
)

type Repository struct {
	pool postgrespool.Pool
}

func NewRepository(pool postgrespool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}
