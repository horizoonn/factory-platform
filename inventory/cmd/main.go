package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"buf.build/go/protovalidate"
	protovalidatemiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	inventoryapi "github.com/horizoonn/factory-platform/inventory/internal/api/inventory/v1"
	"github.com/horizoonn/factory-platform/inventory/internal/config"
	partrepository "github.com/horizoonn/factory-platform/inventory/internal/repository/part"
	partservice "github.com/horizoonn/factory-platform/inventory/internal/service/part"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

func main() {
	if err := run(); err != nil {
		slog.Error("inventory service failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.NewConfigMust()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	postgresPool, err := pgxpool.NewPool(ctx, pgxpool.NewConfigMust())
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer postgresPool.Close()

	postgresRepository := partrepository.NewRepository(postgresPool)
	inventoryService := partservice.NewService(postgresRepository)
	inventoryServer := inventoryapi.NewServer(inventoryService)

	listener, err := net.Listen("tcp", cfg.InventoryGRPC().Address())
	if err != nil {
		return fmt.Errorf("listen inventory grpc address %q: %w", cfg.InventoryGRPC().Address(), err)
	}
	defer func() {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			slog.Warn("failed to close inventory grpc listener", "error", err)
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

	slog.Info("inventory grpc server started", "address", cfg.InventoryGRPC().Address())

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("serve inventory grpc: %w", err)
		}
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownDone := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		slog.Info("inventory grpc server stopped gracefully")
	case <-time.After(cfg.App().ShutdownTimeout()):
		slog.Warn("graceful shutdown timeout exceeded, stopping grpc server forcefully")
		grpcServer.Stop()
	}

	return nil
}
