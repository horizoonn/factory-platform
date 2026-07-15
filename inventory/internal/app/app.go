package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"buf.build/go/protovalidate"
	protovalidatemiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpchealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/horizoonn/factory-platform/inventory/internal/config"
	"github.com/horizoonn/factory-platform/platform/pkg/database/postgres/migrator"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
	"github.com/horizoonn/factory-platform/platform/pkg/swaggerui"
	sharedapi "github.com/horizoonn/factory-platform/shared/api"
	inventoryv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

const readHeaderTimeout = 5 * time.Second

type App struct {
	cfg           config.Config
	diContainer   *diContainer
	grpcServer    *grpc.Server
	httpServer    *http.Server
	health        *grpchealth.Server
	grpcListener  net.Listener
	httpListener  net.Listener
	gatewayCancel context.CancelFunc
}

func New(cfg config.Config) *App {
	return &App{
		cfg:         cfg,
		diContainer: newDIContainer(cfg),
	}
}

func (a *App) Run(ctx context.Context) error {
	if err := a.init(ctx); err != nil {
		a.closeListeners(ctx)
		a.close(ctx)
		return fmt.Errorf("initialize inventory app: %w", err)
	}

	group, runCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return componentError("run inventory grpc server", a.runGRPC(runCtx))
	})
	group.Go(func() error {
		return componentError("run inventory http gateway", a.runHTTP(runCtx))
	})

	logger.Info(ctx, "inventory service started")

	err := group.Wait()
	if ctx.Err() != nil {
		logger.Info(ctx, "shutdown signal received")
	}

	a.closeListeners(ctx)
	a.close(ctx)
	if err != nil {
		return err
	}

	logger.Info(ctx, "inventory service stopped")

	return nil
}

func (a *App) runGRPC(ctx context.Context) error {
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- a.grpcServer.Serve(a.grpcListener)
	}()

	logger.Info(ctx, "inventory grpc server started", zap.String("address", a.cfg.InventoryGRPC().Address()))

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("serve inventory grpc: %w", err)
		}
	case <-ctx.Done():
		a.shutdownGRPCServer(ctx)
	}

	return nil
}

func (a *App) runHTTP(ctx context.Context) error {
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- a.httpServer.Serve(a.httpListener)
	}()

	logger.Info(ctx, "inventory http gateway started", zap.String("address", a.cfg.InventoryHTTP().Address()))

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("serve inventory http gateway: %w", err)
		}
	case <-ctx.Done():
		return a.shutdownHTTPServer(ctx)
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

	if err := a.initGRPCListener(); err != nil {
		return err
	}

	if err := a.initGRPCServer(); err != nil {
		return err
	}

	if err := a.initHTTPListener(); err != nil {
		return err
	}

	return a.initHTTPServer(ctx)
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

func (a *App) initGRPCListener() error {
	listener, err := net.Listen("tcp", a.cfg.InventoryGRPC().Address())
	if err != nil {
		return fmt.Errorf("listen inventory grpc address %q: %w", a.cfg.InventoryGRPC().Address(), err)
	}

	a.grpcListener = listener

	return nil
}

func (a *App) initHTTPListener() error {
	listener, err := net.Listen("tcp", a.cfg.InventoryHTTP().Address())
	if err != nil {
		return fmt.Errorf("listen inventory http address %q: %w", a.cfg.InventoryHTTP().Address(), err)
	}

	a.httpListener = listener

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

func (a *App) initHTTPServer(ctx context.Context) error {
	gatewayCtx, gatewayCancel := context.WithCancel(ctx)
	grpcEndpoint, err := a.grpcGatewayEndpoint()
	if err != nil {
		gatewayCancel()
		return err
	}

	gatewayMux, err := newInventoryGateway(gatewayCtx, grpcEndpoint)
	if err != nil {
		gatewayCancel()
		return err
	}

	a.gatewayCancel = gatewayCancel
	mux := http.NewServeMux()
	mux.Handle("/docs/", swaggerui.NewHandler(swaggerui.Config{
		Title:           "Inventory API",
		UIPath:          "/docs/",
		SpecPath:        "/docs/openapi.json",
		Spec:            sharedapi.InventoryOpenAPIV1(),
		SpecContentType: "application/json",
	}))
	mux.Handle("/", gatewayMux)

	a.httpServer = &http.Server{
		Addr:              a.cfg.InventoryHTTP().Address(),
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	return nil
}

func newInventoryGateway(ctx context.Context, grpcEndpoint string) (*runtime.ServeMux, error) {
	gatewayMux := runtime.NewServeMux()
	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if err := inventoryv1.RegisterInventoryServiceHandlerFromEndpoint(
		ctx,
		gatewayMux,
		grpcEndpoint,
		dialOptions,
	); err != nil {
		return nil, fmt.Errorf("register inventory grpc gateway: %w", err)
	}

	return gatewayMux, nil
}

func (a *App) grpcGatewayEndpoint() (string, error) {
	_, port, err := net.SplitHostPort(a.grpcListener.Addr().String())
	if err != nil {
		return "", fmt.Errorf("split inventory grpc listener address: %w", err)
	}

	return net.JoinHostPort("127.0.0.1", port), nil
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

func (a *App) shutdownHTTPServer(ctx context.Context) error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), a.cfg.App().ShutdownTimeout())
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown inventory http gateway: %w", err)
	}

	logger.Info(ctx, "inventory http gateway stopped gracefully")

	return nil
}

func (a *App) closeListeners(ctx context.Context) {
	closeListener(ctx, a.grpcListener, "inventory grpc")
	closeListener(ctx, a.httpListener, "inventory http")
}

func (a *App) close(ctx context.Context) {
	if a.gatewayCancel != nil {
		a.gatewayCancel()
		a.gatewayCancel = nil
	}

	a.diContainer.Close(ctx)
}

func componentError(component string, err error) error {
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}

	return fmt.Errorf("%s: %w", component, err)
}

func closeListener(ctx context.Context, listener net.Listener, component string) {
	if listener == nil {
		return
	}

	if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		logger.Warn(ctx, "failed to close listener", zap.String("component", component), zap.Error(err))
	}
}

func closeMigrationDB(ctx context.Context, db *sql.DB) {
	if err := db.Close(); err != nil {
		logger.Warn(ctx, "failed to close inventory migration db", zap.Error(err))
	}
}
