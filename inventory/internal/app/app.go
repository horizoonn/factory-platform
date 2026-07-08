package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"buf.build/go/protovalidate"
	protovalidatemiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/horizoonn/factory-platform/inventory/internal/config"
	"github.com/horizoonn/factory-platform/platform/pkg/database/postgres/migrator"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
	inventoryv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

type App struct {
	cfg         config.Config
	diContainer *diContainer
	grpcServer  *grpc.Server
	health      *grpchealth.Server
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
		return fmt.Errorf("initialize inventory app: %w", err)
	}
	defer a.diContainer.Close(ctx)
	defer a.closeListener(ctx)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- a.grpcServer.Serve(a.listener)
	}()

	logger.Info(ctx, "inventory grpc server started", zap.String("address", a.cfg.InventoryGRPC().Address()))

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("serve inventory grpc: %w", err)
		}
	case <-ctx.Done():
		logger.Info(ctx, "shutdown signal received")
	}

	a.shutdownGRPCServer(ctx)

	return nil
}

func (a *App) init(ctx context.Context) error {
	if err := a.diContainer.InitPostgresPool(ctx); err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}

	if err := a.runMigrations(ctx); err != nil {
		return err
	}

	if err := a.initListener(); err != nil {
		return err
	}

	return a.initGRPCServer()
}

func (a *App) runMigrations(ctx context.Context) error {
	db := stdlib.OpenDBFromPool(a.diContainer.PostgresPool().PgxPool())
	defer closeMigrationDB(ctx, db)

	m := migrator.NewMigrator(db, a.cfg.Migrations().Dir())
	if err := m.Up(ctx); err != nil {
		return fmt.Errorf("run inventory migrations: %w", err)
	}

	return nil
}

func (a *App) initListener() error {
	listener, err := net.Listen("tcp", a.cfg.InventoryGRPC().Address())
	if err != nil {
		return fmt.Errorf("listen inventory grpc address %q: %w", a.cfg.InventoryGRPC().Address(), err)
	}

	a.listener = listener

	return nil
}

func (a *App) initGRPCServer() error {
	validator, err := protovalidate.New()
	if err != nil {
		return fmt.Errorf("create protovalidate validator: %w", err)
	}

	a.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(protovalidatemiddleware.UnaryServerInterceptor(validator)),
	)

	inventoryv1.RegisterInventoryServiceServer(a.grpcServer, a.diContainer.InventoryV1API())

	a.health = grpchealth.NewServer()
	healthpb.RegisterHealthServer(a.grpcServer, a.health)
	a.health.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	return nil
}

func (a *App) shutdownGRPCServer(ctx context.Context) {
	if a.health != nil {
		a.health.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	}

	shutdownDone := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		logger.Info(ctx, "inventory grpc server stopped gracefully")
	case <-time.After(a.cfg.App().ShutdownTimeout()):
		logger.Warn(ctx, "graceful shutdown timeout exceeded, stopping grpc server forcefully")
		a.grpcServer.Stop()
	}
}

func (a *App) closeListener(ctx context.Context) {
	if a.listener == nil {
		return
	}

	if err := a.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		logger.Warn(ctx, "failed to close inventory grpc listener", zap.Error(err))
	}
}

func closeMigrationDB(ctx context.Context, db *sql.DB) {
	if err := db.Close(); err != nil {
		logger.Warn(ctx, "failed to close inventory migration db", zap.Error(err))
	}
}
