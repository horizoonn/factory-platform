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

	paymentapi "github.com/horizoonn/factory-platform.git/payment/internal/api/payment/v1"
	"github.com/horizoonn/factory-platform.git/payment/internal/config"
	"github.com/horizoonn/factory-platform.git/payment/internal/service"
	paymentv1 "github.com/horizoonn/factory-platform.git/shared/pkg/proto/payment/v1"
)

func main() {
	if err := run(); err != nil {
		slog.Error("payment service failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.NewConfigMust()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	paymentService := service.NewPaymentService()
	paymentServer := paymentapi.NewPaymentServer(paymentService)

	listener, err := net.Listen("tcp", cfg.PaymentGRPC().Address())
	if err != nil {
		return fmt.Errorf("listen payment grpc address %q: %w", cfg.PaymentGRPC().Address(), err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			slog.Warn("failed to close payment grpc listener", "error", err)
		}
	}()

	grpcServer := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(grpcServer, paymentServer)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- grpcServer.Serve(listener)
	}()

	slog.Info("payment grpc server started", "address", cfg.PaymentGRPC().Address())

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("serve payment grpc: %w", err)
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
		slog.Info("payment grpc server stopped gracefully")
	case <-time.After(cfg.App().ShutdownTimeout()):
		slog.Warn("graceful shutdown timeout exceeded, stopping grpc server forcefully")
		grpcServer.Stop()
	}

	return nil
}
