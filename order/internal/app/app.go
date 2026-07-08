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

	"github.com/horizoonn/factory-platform/order/internal/config"
	"github.com/horizoonn/factory-platform/platform/pkg/database/postgres/migrator"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
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
		a.diContainer.Close(ctx)
		return fmt.Errorf("initialize order app: %w", err)
	}
	defer a.diContainer.Close(ctx)
	defer a.closeListener(ctx)

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
		logger.Info(ctx, "shutdown signal received")
	}

	if err := a.shutdownHTTPServer(ctx); err != nil {
		return err
	}

	return nil
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

func closeMigrationDB(ctx context.Context, db *sql.DB) {
	if err := db.Close(); err != nil {
		logger.Warn(ctx, "failed to close order migration db", zap.Error(err))
	}
}
