package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/horizoonn/factory-platform/order/internal/api/health"
	shipassembledapi "github.com/horizoonn/factory-platform/order/internal/api/kafka/shipassembled"
	orderapi "github.com/horizoonn/factory-platform/order/internal/api/order/v1"
	grpcclients "github.com/horizoonn/factory-platform/order/internal/client/grpc"
	"github.com/horizoonn/factory-platform/order/internal/config"
	outboxdispatcher "github.com/horizoonn/factory-platform/order/internal/outbox/dispatcher"
	"github.com/horizoonn/factory-platform/order/internal/outbox/encoder"
	orderrepository "github.com/horizoonn/factory-platform/order/internal/repository/order"
	outboxrepository "github.com/horizoonn/factory-platform/order/internal/repository/outbox"
	orderservice "github.com/horizoonn/factory-platform/order/internal/service/order"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer"
	consumerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer/franz"
	producerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/producer/franz"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
	kafkamiddleware "github.com/horizoonn/factory-platform/platform/pkg/middleware/kafka"
	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

type diContainer struct {
	cfg config.Config

	postgresPool          *pgxpool.Pool
	grpcClients           *grpcclients.Clients
	orderPaidProducer     *producerfranz.Producer
	shipAssembledConsumer *consumerfranz.Consumer

	outboxRepository *outboxrepository.Repository
	orderRepository  orderservice.Repository

	orderPaidEncoder *encoder.OrderPaid
	orderService     *orderservice.Service

	shipAssembledHandler consumer.Handler
	orderHandler         orderopenapi.Handler
	healthChecker        *health.Checker

	outboxDispatcher *outboxdispatcher.Dispatcher
	orderHTTPServer  http.Handler
}

func newDIContainer(cfg config.Config) *diContainer {
	return &diContainer{cfg: cfg}
}

func (d *diContainer) InitPostgresPool(ctx context.Context) error {
	if d.postgresPool != nil {
		return nil
	}

	postgresPool, err := pgxpool.NewPool(ctx, d.cfg.Postgres())
	if err != nil {
		return err
	}

	d.postgresPool = postgresPool
	return nil
}

func (d *diContainer) InitGRPCClients() error {
	if d.grpcClients != nil {
		return nil
	}

	grpcClients, err := grpcclients.NewClients(
		d.cfg.InventoryGRPC().Address(),
		d.cfg.PaymentGRPC().Address(),
	)
	if err != nil {
		return err
	}

	d.grpcClients = grpcClients
	return nil
}

func (d *diContainer) InitOrderPaidProducer(ctx context.Context) error {
	if d.orderPaidProducer != nil {
		return nil
	}

	kafkaProducer, err := producerfranz.NewProducer(d.cfg.OrderPaidProducer())
	if err != nil {
		return err
	}

	if err := kafkaProducer.Ping(ctx); err != nil {
		return errors.Join(err, kafkaProducer.Close(ctx))
	}

	d.orderPaidProducer = kafkaProducer
	return nil
}

func (d *diContainer) InitShipAssembledConsumer(ctx context.Context) error {
	if d.shipAssembledConsumer != nil {
		return nil
	}

	kafkaConsumer, err := consumerfranz.NewConsumer(
		d.cfg.ShipAssembledConsumer(),
		consumerfranz.WithLogger(logger.Default()),
		consumerfranz.WithMiddlewares(kafkamiddleware.Logging(logger.Default())),
	)
	if err != nil {
		return err
	}

	if err := kafkaConsumer.Ping(ctx); err != nil {
		return errors.Join(err, kafkaConsumer.Close(ctx))
	}

	d.shipAssembledConsumer = kafkaConsumer
	return nil
}

func (d *diContainer) InitOutboxDispatcher() error {
	if d.outboxDispatcher != nil {
		return nil
	}
	if d.orderPaidProducer == nil {
		return errors.New("kafka producer is not initialized")
	}

	dispatcherConfig := d.cfg.OutboxDispatcher()
	workerID, err := newOutboxWorkerID()
	if err != nil {
		return err
	}
	dispatcherConfig.WorkerID = workerID

	dispatcher, err := outboxdispatcher.NewDispatcher(
		d.OutboxRepository(),
		d.orderPaidProducer,
		dispatcherConfig,
	)
	if err != nil {
		return err
	}

	d.outboxDispatcher = dispatcher
	return nil
}

func (d *diContainer) PostgresPool() *pgxpool.Pool {
	return d.postgresPool
}

func (d *diContainer) ShipAssembledConsumer() *consumerfranz.Consumer {
	return d.shipAssembledConsumer
}

func (d *diContainer) OutboxDispatcher() *outboxdispatcher.Dispatcher {
	return d.outboxDispatcher
}

func (d *diContainer) OutboxRepository() *outboxrepository.Repository {
	if d.outboxRepository == nil {
		if d.postgresPool == nil {
			panic("postgres pool is not initialized")
		}

		d.outboxRepository = outboxrepository.NewRepository(d.postgresPool)
	}

	return d.outboxRepository
}

func (d *diContainer) OrderRepository() orderservice.Repository {
	if d.orderRepository == nil {
		if d.postgresPool == nil {
			panic("postgres pool is not initialized")
		}

		d.orderRepository = orderrepository.NewRepository(
			d.postgresPool,
			d.OutboxRepository(),
		)
	}

	return d.orderRepository
}

func (d *diContainer) OrderPaidEncoder() *encoder.OrderPaid {
	if d.orderPaidEncoder == nil {
		d.orderPaidEncoder = encoder.NewOrderPaid()
	}

	return d.orderPaidEncoder
}

func (d *diContainer) OrderService() *orderservice.Service {
	if d.orderService == nil {
		if d.grpcClients == nil {
			panic("grpc clients are not initialized")
		}

		d.orderService = orderservice.NewService(
			d.OrderRepository(),
			d.grpcClients.Inventory,
			d.grpcClients.Payment,
			d.OrderPaidEncoder(),
		)
	}

	return d.orderService
}

func (d *diContainer) ShipAssembledHandler() consumer.Handler {
	if d.shipAssembledHandler == nil {
		d.shipAssembledHandler = shipassembledapi.NewHandler(d.OrderService())
	}

	return d.shipAssembledHandler
}

func (d *diContainer) OrderHandler() orderopenapi.Handler {
	if d.orderHandler == nil {
		d.orderHandler = orderapi.NewHandler(d.OrderService())
	}

	return d.orderHandler
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

func (d *diContainer) Close(ctx context.Context) {
	if d.shipAssembledConsumer != nil {
		if err := d.shipAssembledConsumer.Close(ctx); err != nil {
			logger.Warn(ctx, "failed to close kafka consumer", zap.Error(err))
		}
		d.shipAssembledConsumer = nil
	}

	if d.orderPaidProducer != nil {
		if err := d.orderPaidProducer.Close(ctx); err != nil {
			logger.Warn(ctx, "failed to close kafka producer", zap.Error(err))
		}
		d.orderPaidProducer = nil
	}

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

func newOutboxWorkerID() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("get hostname for outbox worker id: %w", err)
	}

	return hostname + "-" + uuid.NewString(), nil
}
