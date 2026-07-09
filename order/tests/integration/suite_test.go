//go:build integration

package integration

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/platform/pkg/database/postgres/migrator"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	tcpostgres "github.com/horizoonn/factory-platform/platform/pkg/testcontainers/postgres"
)

const testTimeout = time.Minute

var testEnv *environment

type environment struct {
	container *tcpostgres.Container
	pool      *pgxpool.Pool
	db        *sql.DB
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	env, err := newEnvironment(ctx)
	if err != nil {
		panic(err)
	}
	testEnv = env

	code := m.Run()

	if err := env.close(context.Background()); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func newEnvironment(ctx context.Context) (*environment, error) {
	container, err := tcpostgres.NewContainer(
		ctx,
		tcpostgres.WithDatabase("order"),
		tcpostgres.WithUsername("order"),
		tcpostgres.WithPassword("order"),
	)
	if err != nil {
		return nil, err
	}

	success := false
	defer func() {
		if !success {
			_ = container.Terminate(ctx)
		}
	}()

	poolConfig, err := container.PgxPoolConfig(ctx)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewPool(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	db := stdlib.OpenDBFromPool(pool.PgxPool())
	if err := migrator.NewMigrator(db, migrationsDir()).Up(ctx); err != nil {
		db.Close()
		pool.Close()
		return nil, err
	}

	success = true

	return &environment{
		container: container,
		pool:      pool,
		db:        db,
	}, nil
}

func (e *environment) close(ctx context.Context) error {
	var err error
	if e.db != nil {
		err = errors.Join(err, e.db.Close())
	}
	if e.pool != nil {
		e.pool.Close()
	}
	if e.container != nil {
		err = errors.Join(err, e.container.Terminate(ctx))
	}

	return err
}

func (e *environment) truncateOrders(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.pool.Exec(ctx, "TRUNCATE TABLE platform.orders CASCADE")
	require.NoError(t, err)
}

func testContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	return ctx
}

func migrationsDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("get current test filename")
	}

	return filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
}
