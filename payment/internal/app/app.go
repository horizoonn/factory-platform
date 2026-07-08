package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"buf.build/go/protovalidate"
	protovalidatemiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/horizoonn/factory-platform/payment/internal/config"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
	paymentv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
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
		diContainer: newDIContainer(),
	}
}

func (a *App) Run(ctx context.Context) error {
	if err := a.init(); err != nil {
		a.closeListener(ctx)
		return fmt.Errorf("initialize payment app: %w", err)
	}
	defer a.closeListener(ctx)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- a.grpcServer.Serve(a.listener)
	}()

	logger.Info(ctx, "payment grpc server started", zap.String("address", a.cfg.PaymentGRPC().Address()))

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("serve payment grpc: %w", err)
		}
	case <-ctx.Done():
		logger.Info(ctx, "shutdown signal received")
	}

	a.shutdownGRPCServer(ctx)

	return nil
}

func (a *App) init() error {
	if err := a.initListener(); err != nil {
		return err
	}

	return a.initGRPCServer()
}

func (a *App) initListener() error {
	listener, err := net.Listen("tcp", a.cfg.PaymentGRPC().Address())
	if err != nil {
		return fmt.Errorf("listen payment grpc address %q: %w", a.cfg.PaymentGRPC().Address(), err)
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

	paymentv1.RegisterPaymentServiceServer(a.grpcServer, a.diContainer.PaymentV1API())

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
		logger.Info(ctx, "payment grpc server stopped gracefully")
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
		logger.Warn(ctx, "failed to close payment grpc listener", zap.Error(err))
	}
}
