package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/horizoonn/factory-platform/order/internal/config"
	"github.com/horizoonn/factory-platform/platform/pkg/database/postgres/migrator"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
	"github.com/horizoonn/factory-platform/platform/pkg/swaggerui"
	sharedapi "github.com/horizoonn/factory-platform/shared/api"
)

const readHeaderTimeout = 5 * time.Second

type App struct {
	cfg         config.Config
	diContainer *diContainer
	httpServer  *http.Server
	listener    net.Listener
}

func New(cfg config.Config) *App {
	return &App{
		cfg:         cfg,
		diContainer: newDIContainer(cfg),
	}
}

func (a *App) Run(ctx context.Context) error {
	if err := a.init(ctx); err != nil {
		a.closeListener(ctx)
		a.close(ctx)
		return fmt.Errorf("initialize order app: %w", err)
	}

	group, runCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return componentError("run order http server", a.runHTTP(runCtx))
	})
	group.Go(func() error {
		return componentError("run outbox dispatcher", a.diContainer.OutboxDispatcher().Run(runCtx))
	})
	group.Go(func() error {
		err := a.diContainer.ShipAssembledConsumer().Consume(
			runCtx,
			a.diContainer.ShipAssembledHandler(),
		)
		return componentError("consume ShipAssembled events", err)
	})

	logger.Info(ctx, "order service started")

	err := group.Wait()
	if ctx.Err() != nil {
		logger.Info(ctx, "shutdown signal received")
	}

	a.closeListener(ctx)
	a.close(ctx)
	if err != nil {
		return err
	}

	logger.Info(ctx, "order service stopped")

	return nil
}

func (a *App) runHTTP(ctx context.Context) error {
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- a.httpServer.Serve(a.listener)
	}()

	logger.Info(ctx, "order http server started", zap.String("address", a.cfg.OrderHTTP().Address()))

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("serve order http: %w", err)
		}
	case <-ctx.Done():
	}

	return a.shutdownHTTPServer(ctx)
}

func (a *App) init(ctx context.Context) error {
	if err := a.diContainer.InitPostgresPool(ctx); err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}

	if err := a.runMigrations(ctx); err != nil {
		return err
	}

	if err := a.diContainer.InitGRPCClients(); err != nil {
		return fmt.Errorf("create grpc clients: %w", err)
	}

	if err := a.diContainer.InitOrderPaidProducer(ctx); err != nil {
		return fmt.Errorf("create OrderPaid kafka producer: %w", err)
	}

	if err := a.diContainer.InitShipAssembledConsumer(ctx); err != nil {
		return fmt.Errorf("create ShipAssembled kafka consumer: %w", err)
	}

	if err := a.diContainer.InitOutboxDispatcher(); err != nil {
		return fmt.Errorf("create outbox dispatcher: %w", err)
	}

	if err := a.initHTTPServer(); err != nil {
		return err
	}

	return a.initListener()
}

func (a *App) runMigrations(ctx context.Context) error {
	db := stdlib.OpenDBFromPool(a.diContainer.PostgresPool().PgxPool())
	defer closeMigrationDB(ctx, db)

	m := migrator.NewMigrator(db, a.cfg.Migrations().Dir())
	if err := m.Up(ctx); err != nil {
		return fmt.Errorf("run order migrations: %w", err)
	}

	return nil
}

func (a *App) initHTTPServer() error {
	orderServer, err := a.diContainer.OrderHTTPServer()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", a.diContainer.HealthChecker().Handler)
	mux.Handle("/docs/", swaggerui.NewHandler(swaggerui.Config{
		Title:           "Order API",
		UIPath:          "/docs/",
		SpecPath:        "/docs/openapi.yaml",
		Spec:            sharedapi.OrderOpenAPIV1(),
		SpecContentType: "application/yaml",
	}))
	mux.Handle("/", orderServer)

	a.httpServer = &http.Server{
		Addr:              a.cfg.OrderHTTP().Address(),
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	return nil
}

func (a *App) initListener() error {
	listener, err := net.Listen("tcp", a.cfg.OrderHTTP().Address())
	if err != nil {
		return fmt.Errorf("listen order http address %q: %w", a.cfg.OrderHTTP().Address(), err)
	}

	a.listener = listener

	return nil
}

func (a *App) shutdownHTTPServer(ctx context.Context) error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), a.cfg.App().ShutdownTimeout())
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown order http server: %w", err)
	}

	logger.Info(ctx, "order http server stopped gracefully")

	return nil
}

func (a *App) closeListener(ctx context.Context) {
	if a.listener == nil {
		return
	}

	if err := a.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		logger.Warn(ctx, "failed to close order http listener", zap.Error(err))
	}
}

func (a *App) close(ctx context.Context) {
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		a.cfg.App().ShutdownTimeout(),
	)
	defer shutdownCancel()

	a.diContainer.Close(shutdownCtx)
}

func componentError(component string, err error) error {
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}

	return fmt.Errorf("%s: %w", component, err)
}

func closeMigrationDB(ctx context.Context, db *sql.DB) {
	if err := db.Close(); err != nil {
		logger.Warn(ctx, "failed to close order migration db", zap.Error(err))
	}
}
