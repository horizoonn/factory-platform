package app

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/horizoonn/factory-platform/order/internal/api/health"
	orderapi "github.com/horizoonn/factory-platform/order/internal/api/order/v1"
	grpcclients "github.com/horizoonn/factory-platform/order/internal/client/grpc"
	"github.com/horizoonn/factory-platform/order/internal/config"
	orderrepository "github.com/horizoonn/factory-platform/order/internal/repository/order"
	orderservice "github.com/horizoonn/factory-platform/order/internal/service/order"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

type diContainer struct {
	cfg config.Config

	orderHTTPServer http.Handler
	orderHandler    orderopenapi.Handler
	orderService    orderapi.OrderService
	orderRepository orderservice.Repository
	healthChecker   *health.Checker

	postgresPool *pgxpool.Pool
	grpcClients  *grpcclients.Clients
}

func newDIContainer(cfg config.Config) *diContainer {
	return &diContainer{
		cfg: cfg,
	}
}

func (d *diContainer) OrderHTTPServer() (http.Handler, error) {
	if d.orderHTTPServer == nil {
		orderServer, err := orderopenapi.NewServer(d.OrderHandler())
		if err != nil {
			return nil, fmt.Errorf("create order openapi server: %w", err)
		}

		d.orderHTTPServer = orderServer
	}

	return d.orderHTTPServer, nil
}

func (d *diContainer) OrderHandler() orderopenapi.Handler {
	if d.orderHandler == nil {
		d.orderHandler = orderapi.NewHandler(d.OrderService())
	}

	return d.orderHandler
}

func (d *diContainer) OrderService() orderapi.OrderService {
	if d.orderService == nil {
		if d.grpcClients == nil {
			panic("grpc clients are not initialized")
		}
		d.orderService = orderservice.NewService(
			d.OrderRepository(),
			d.grpcClients.Inventory,
			d.grpcClients.Payment,
		)
	}

	return d.orderService
}

func (d *diContainer) OrderRepository() orderservice.Repository {
	if d.orderRepository == nil {
		if d.postgresPool == nil {
			panic("postgres pool is not initialized")
		}
		d.orderRepository = orderrepository.NewRepository(d.postgresPool)
	}

	return d.orderRepository
}

func (d *diContainer) HealthChecker() *health.Checker {
	if d.healthChecker == nil {
		if d.postgresPool == nil {
			panic("postgres pool is not initialized")
		}
		if d.grpcClients == nil {
			panic("grpc clients are not initialized")
		}
		d.healthChecker = health.NewChecker(d.postgresPool, d.grpcClients.Connections())
	}

	return d.healthChecker
}

func (d *diContainer) InitPostgresPool(ctx context.Context) error {
	if d.postgresPool == nil {
		postgresPool, err := pgxpool.NewPool(ctx, d.cfg.Postgres())
		if err != nil {
			return err
		}

		d.postgresPool = postgresPool
	}

	return nil
}

func (d *diContainer) InitGRPCClients() error {
	if d.grpcClients == nil {
		grpcClients, err := grpcclients.NewClients(
			d.cfg.InventoryGRPC().Address(),
			d.cfg.PaymentGRPC().Address(),
		)
		if err != nil {
			return err
		}

		d.grpcClients = grpcClients
	}

	return nil
}

func (d *diContainer) PostgresPool() *pgxpool.Pool {
	return d.postgresPool
}

func (d *diContainer) Close(ctx context.Context) {
	if d.grpcClients != nil {
		if err := d.grpcClients.Close(); err != nil {
			logger.Warn(ctx, "failed to close grpc clients", zap.Error(err))
		}
		d.grpcClients = nil
	}

	if d.postgresPool != nil {
		d.postgresPool.Close()
		d.postgresPool = nil
	}
}
