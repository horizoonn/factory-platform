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

	"google.golang.org/grpc"

	inventoryapi "github.com/horizoonn/factory-platform.git/inventory/internal/api/inventory/v1"
	"github.com/horizoonn/factory-platform.git/inventory/internal/config"
	repository "github.com/horizoonn/factory-platform.git/inventory/internal/repository/part"
	"github.com/horizoonn/factory-platform.git/inventory/internal/service"
	pgxpool "github.com/horizoonn/factory-platform.git/platform/pkg/database/postgres/pool/pgx"
	inventorypb "github.com/horizoonn/factory-platform.git/shared/pkg/proto/inventory/v1"
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

	postgresRepository := repository.NewRepository(postgresPool)
	inventoryService := service.NewInventoryService(postgresRepository)
	inventoryServer := inventoryapi.NewInventoryServer(inventoryService)

	listener, err := net.Listen("tcp", cfg.InventoryGRPC().Address())
	if err != nil {
		return fmt.Errorf("listen inventory grpc address %q: %w", cfg.InventoryGRPC().Address(), err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			slog.Warn("failed to close inventory grpc listener", "error", err)
		}
	}()

	grpcServer := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(grpcServer, inventoryServer)

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
