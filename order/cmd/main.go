package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/horizoonn/factory-platform/order/internal/api/health"
	orderapi "github.com/horizoonn/factory-platform/order/internal/api/order/v1"
	inventoryclient "github.com/horizoonn/factory-platform/order/internal/client/grpc/inventory/v1"
	paymentclient "github.com/horizoonn/factory-platform/order/internal/client/grpc/payment/v1"
	"github.com/horizoonn/factory-platform/order/internal/config"
	repository "github.com/horizoonn/factory-platform/order/internal/repository/order"
	"github.com/horizoonn/factory-platform/order/internal/service"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

func main() {
	if err := run(); err != nil {
		slog.Error("order service failed", "error", err)
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

	inventoryConn, err := grpc.NewClient(cfg.InventoryGRPC().Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create inventory grpc client: %w", err)
	}
	defer func() {
		if err := inventoryConn.Close(); err != nil {
			slog.Warn("failed to close inventory grpc connection", "error", err)
		}
	}()

	paymentConn, err := grpc.NewClient(cfg.PaymentGRPC().Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create payment grpc client: %w", err)
	}
	defer func() {
		if err := paymentConn.Close(); err != nil {
			slog.Warn("failed to close payment grpc connection", "error", err)
		}
	}()

	orderRepository := repository.NewRepository(postgresPool)
	inventoryClient := inventoryclient.NewClient(inventoryConn)
	paymentClient := paymentclient.NewClient(paymentConn)
	orderService := service.NewOrderService(orderRepository, inventoryClient, paymentClient)
	orderHandler := orderapi.NewOrderHandler(orderService)

	orderServer, err := orderopenapi.NewServer(orderHandler)
	if err != nil {
		return fmt.Errorf("create order openapi server: %w", err)
	}

	healthChecker := health.NewChecker(postgresPool, map[string]*grpc.ClientConn{
		"inventory": inventoryConn,
		"payment":   paymentConn,
	})

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
			slog.Warn("failed to close order http listener", "error", err)
		}
	}()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpServer.Serve(listener)
	}()

	slog.Info("order http server started", "address", cfg.OrderHTTP().Address())

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve order http: %w", err)
		}
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.App().ShutdownTimeout())
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown order http server: %w", err)
	}

	slog.Info("order http server stopped gracefully")

	return nil
}
