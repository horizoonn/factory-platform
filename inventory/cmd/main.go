package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"buf.build/go/protovalidate"
	protovalidatemiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	inventoryapi "github.com/horizoonn/factory-platform/inventory/internal/api/inventory/v1"
	"github.com/horizoonn/factory-platform/inventory/internal/config"
	partrepository "github.com/horizoonn/factory-platform/inventory/internal/repository/part"
	partservice "github.com/horizoonn/factory-platform/inventory/internal/service/part"
	"github.com/horizoonn/factory-platform/platform/pkg/database/postgres/migrator"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

func main() {
	if err := run(); err != nil {
		logger.Error(context.Background(), "inventory service failed", zap.Error(err))
		fmt.Fprintf(os.Stderr, "inventory service failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.NewConfigMust()

	if err := logger.InitWithConfig(cfg.Logger()); err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Warn(context.Background(), "failed to sync logger", zap.Error(err))
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	postgresPool, err := pgxpool.NewPool(ctx, pgxpool.NewConfigMust())
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer postgresPool.Close()

	db := stdlib.OpenDBFromPool(postgresPool.Pool)
	defer func() {
		if err := db.Close(); err != nil {
			logger.Warn(context.Background(), "failed to close inventory migration db", zap.Error(err))
		}
	}()
	m := migrator.NewMigrator(db, cfg.Migrations().Dir())
	if err := m.Up(ctx); err != nil {
		return fmt.Errorf("run inventory migrations: %w", err)
	}

	postgresRepository := partrepository.NewRepository(postgresPool)
	inventoryService := partservice.NewService(postgresRepository)
	inventoryServer := inventoryapi.NewServer(inventoryService)

	listener, err := net.Listen("tcp", cfg.InventoryGRPC().Address())
	if err != nil {
		return fmt.Errorf("listen inventory grpc address %q: %w", cfg.InventoryGRPC().Address(), err)
	}
	defer func() {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Warn(context.Background(), "failed to close inventory grpc listener", zap.Error(err))
		}
	}()

	validator, err := protovalidate.New()
	if err != nil {
		return fmt.Errorf("create protovalidate validator: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(protovalidatemiddleware.UnaryServerInterceptor(validator)),
	)
	inventorypb.RegisterInventoryServiceServer(grpcServer, inventoryServer)
	healthServer := grpchealth.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- grpcServer.Serve(listener)
	}()

	logger.Info(ctx, "inventory grpc server started", zap.String("address", cfg.InventoryGRPC().Address()))

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("serve inventory grpc: %w", err)
		}
	case <-ctx.Done():
		logger.Info(ctx, "shutdown signal received")
	}

	shutdownDone := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		logger.Info(ctx, "inventory grpc server stopped gracefully")
	case <-time.After(cfg.App().ShutdownTimeout()):
		logger.Warn(ctx, "graceful shutdown timeout exceeded, stopping grpc server forcefully")
		grpcServer.Stop()
	}

	return nil
}
