package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/horizoonn/factory-platform/assembly/internal/config"
	"github.com/horizoonn/factory-platform/platform/pkg/database/postgres/migrator"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type App struct {
	cfg         config.Config
	diContainer *diContainer
}

func New(cfg config.Config) *App {
	return &App{
		cfg:         cfg,
		diContainer: newDIContainer(cfg),
	}
}

func (a *App) Run(ctx context.Context) error {
	if err := a.init(ctx); err != nil {
		a.close(ctx)
		return fmt.Errorf("initialize assembly app: %w", err)
	}

	group, runCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		err := a.diContainer.OrderPaidConsumer().Consume(runCtx, a.diContainer.OrderPaidHandler())
		return componentError("consume OrderPaid events", err)
	})
	group.Go(func() error {
		err := a.diContainer.OutboxDispatcher().Run(runCtx)
		return componentError("run outbox dispatcher", err)
	})

	logger.Info(ctx, "assembly service started")

	err := group.Wait()
	if ctx.Err() != nil {
		logger.Info(ctx, "shutdown signal received")
	}

	a.close(ctx)
	if err != nil {
		return err
	}

	logger.Info(ctx, "assembly service stopped")

	return nil
}

func (a *App) init(ctx context.Context) error {
	if err := a.diContainer.InitPostgresPool(ctx); err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}

	if err := a.runMigrations(ctx); err != nil {
		return err
	}

	if err := a.diContainer.InitOrderPaidConsumer(ctx); err != nil {
		return fmt.Errorf("create OrderPaid kafka consumer: %w", err)
	}

	if err := a.diContainer.InitShipAssembledProducer(ctx); err != nil {
		return fmt.Errorf("create ShipAssembled kafka producer: %w", err)
	}

	if err := a.diContainer.InitOutboxDispatcher(); err != nil {
		return fmt.Errorf("create outbox dispatcher: %w", err)
	}

	return nil
}

func (a *App) close(ctx context.Context) {
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		a.cfg.App().ShutdownTimeout(),
	)
	defer shutdownCancel()

	a.diContainer.Close(shutdownCtx)
}

func (a *App) runMigrations(ctx context.Context) error {
	db := stdlib.OpenDBFromPool(a.diContainer.PostgresPool().PgxPool())
	defer closeMigrationDB(ctx, db)

	m := migrator.NewMigrator(db, a.cfg.Migrations().Dir())
	if err := m.Up(ctx); err != nil {
		return fmt.Errorf("run assembly migrations: %w", err)
	}

	return nil
}

func componentError(component string, err error) error {
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}

	return fmt.Errorf("%s: %w", component, err)
}

func closeMigrationDB(ctx context.Context, db *sql.DB) {
	if err := db.Close(); err != nil {
		logger.Warn(ctx, "failed to close assembly migration db", zap.Error(err))
	}
}
