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
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	paymentapi "github.com/horizoonn/factory-platform/payment/internal/api/payment/v1"
	"github.com/horizoonn/factory-platform/payment/internal/config"
	paymentservice "github.com/horizoonn/factory-platform/payment/internal/service/payment"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
	paymentv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
)

func main() {
	if err := run(); err != nil {
		logger.Error(context.Background(), "payment service failed", zap.Error(err))
		fmt.Fprintf(os.Stderr, "payment service failed: %v\n", err)
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

	paymentService := paymentservice.NewService()
	paymentServer := paymentapi.NewServer(paymentService)

	listener, err := net.Listen("tcp", cfg.PaymentGRPC().Address())
	if err != nil {
		return fmt.Errorf("listen payment grpc address %q: %w", cfg.PaymentGRPC().Address(), err)
	}
	defer func() {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Warn(context.Background(), "failed to close payment grpc listener", zap.Error(err))
		}
	}()

	validator, err := protovalidate.New()
	if err != nil {
		return fmt.Errorf("create protovalidate validator: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(protovalidatemiddleware.UnaryServerInterceptor(validator)),
	)
	paymentv1.RegisterPaymentServiceServer(grpcServer, paymentServer)
	healthServer := grpchealth.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- grpcServer.Serve(listener)
	}()

	logger.Info(ctx, "payment grpc server started", zap.String("address", cfg.PaymentGRPC().Address()))

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("serve payment grpc: %w", err)
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
		logger.Info(ctx, "payment grpc server stopped gracefully")
	case <-time.After(cfg.App().ShutdownTimeout()):
		logger.Warn(ctx, "graceful shutdown timeout exceeded, stopping grpc server forcefully")
		grpcServer.Stop()
	}

	return nil
}
