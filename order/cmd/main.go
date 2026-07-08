package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/horizoonn/factory-platform/order/internal/api/health"
	orderapi "github.com/horizoonn/factory-platform/order/internal/api/order/v1"
	grpcclients "github.com/horizoonn/factory-platform/order/internal/client/grpc"
	"github.com/horizoonn/factory-platform/order/internal/config"
	orderrepository "github.com/horizoonn/factory-platform/order/internal/repository/order"
	orderservice "github.com/horizoonn/factory-platform/order/internal/service/order"
	"github.com/horizoonn/factory-platform/platform/pkg/database/postgres/migrator"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

func main() {
	if err := run(); err != nil {
		logger.Error(context.Background(), "order service failed", zap.Error(err))
		fmt.Fprintf(os.Stderr, "order service failed: %v\n", err)
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
			logger.Warn(context.Background(), "failed to close order migration db", zap.Error(err))
		}
	}()
	m := migrator.NewMigrator(db, cfg.Migrations().Dir())
	if err := m.Up(ctx); err != nil {
		return fmt.Errorf("run order migrations: %w", err)
	}

	grpcClients, err := grpcclients.NewClients(cfg.InventoryGRPC().Address(), cfg.PaymentGRPC().Address())
	if err != nil {
		return fmt.Errorf("create grpc clients: %w", err)
	}
	defer func() {
		if err := grpcClients.Close(); err != nil {
			logger.Warn(context.Background(), "failed to close grpc clients", zap.Error(err))
		}
	}()

	orderRepository := orderrepository.NewRepository(postgresPool)
	orderService := orderservice.NewService(orderRepository, grpcClients.Inventory, grpcClients.Payment)
	orderHandler := orderapi.NewHandler(orderService)

	orderServer, err := orderopenapi.NewServer(orderHandler)
	if err != nil {
		return fmt.Errorf("create order openapi server: %w", err)
	}

	healthChecker := health.NewChecker(postgresPool, grpcClients.Connections())

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthChecker.Handler)
	mux.Handle("/", orderServer)

	httpServer := &http.Server{
		Addr:              cfg.OrderHTTP().Address(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	listener, err := net.Listen("tcp", cfg.OrderHTTP().Address())
	if err != nil {
		return fmt.Errorf("listen order http address %q: %w", cfg.OrderHTTP().Address(), err)
	}
	defer func() {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Warn(context.Background(), "failed to close order http listener", zap.Error(err))
		}
	}()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpServer.Serve(listener)
	}()

	logger.Info(ctx, "order http server started", zap.String("address", cfg.OrderHTTP().Address()))

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve order http: %w", err)
		}
	case <-ctx.Done():
		logger.Info(ctx, "shutdown signal received")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.App().ShutdownTimeout())
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown order http server: %w", err)
	}

	logger.Info(ctx, "order http server stopped gracefully")

	return nil
}
