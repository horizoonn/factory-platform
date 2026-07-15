package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"

	orderpaidapi "github.com/horizoonn/factory-platform/assembly/internal/api/kafka/orderpaid"
	"github.com/horizoonn/factory-platform/assembly/internal/config"
	outboxdispatcher "github.com/horizoonn/factory-platform/assembly/internal/outbox/dispatcher"
	"github.com/horizoonn/factory-platform/assembly/internal/outbox/encoder"
	outboxrepository "github.com/horizoonn/factory-platform/assembly/internal/repository/outbox"
	assemblyservice "github.com/horizoonn/factory-platform/assembly/internal/service/assembly"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer"
	consumerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer/franz"
	producerfranz "github.com/horizoonn/factory-platform/platform/pkg/kafka/producer/franz"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
	kafkamiddleware "github.com/horizoonn/factory-platform/platform/pkg/middleware/kafka"
)

type diContainer struct {
	cfg config.Config

	postgresPool          *pgxpool.Pool
	orderPaidConsumer     *consumerfranz.Consumer
	shipAssembledProducer *producerfranz.Producer

	outboxRepository     *outboxrepository.Repository
	shipAssembledEncoder *encoder.ShipAssembled
	assemblyService      *assemblyservice.Service
	orderPaidHandler     consumer.Handler
	outboxDispatcher     *outboxdispatcher.Dispatcher
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

func (d *diContainer) InitOrderPaidConsumer(ctx context.Context) error {
	if d.orderPaidConsumer != nil {
		return nil
	}

	kafkaConsumer, err := consumerfranz.NewConsumer(
		d.cfg.OrderPaidConsumer(),
		consumerfranz.WithLogger(logger.Default()),
		consumerfranz.WithMiddlewares(kafkamiddleware.Logging(logger.Default())),
	)
	if err != nil {
		return err
	}

	if err := kafkaConsumer.Ping(ctx); err != nil {
		return errors.Join(err, kafkaConsumer.Close(ctx))
	}

	d.orderPaidConsumer = kafkaConsumer
	return nil
}

func (d *diContainer) InitShipAssembledProducer(ctx context.Context) error {
	if d.shipAssembledProducer != nil {
		return nil
	}

	kafkaProducer, err := producerfranz.NewProducer(d.cfg.ShipAssembledProducer())
	if err != nil {
		return err
	}

	if err := kafkaProducer.Ping(ctx); err != nil {
		return errors.Join(err, kafkaProducer.Close(ctx))
	}

	d.shipAssembledProducer = kafkaProducer
	return nil
}

func (d *diContainer) PostgresPool() *pgxpool.Pool {
	return d.postgresPool
}

func (d *diContainer) OrderPaidConsumer() *consumerfranz.Consumer {
	return d.orderPaidConsumer
}

func (d *diContainer) OrderPaidHandler() consumer.Handler {
	if d.orderPaidHandler == nil {
		d.orderPaidHandler = orderpaidapi.NewHandler(d.AssemblyService())
	}

	return d.orderPaidHandler
}

func (d *diContainer) AssemblyService() *assemblyservice.Service {
	if d.assemblyService == nil {
		d.assemblyService = assemblyservice.NewService(
			d.OutboxRepository(),
			d.ShipAssembledEncoder(),
		)
	}

	return d.assemblyService
}

func (d *diContainer) ShipAssembledEncoder() *encoder.ShipAssembled {
	if d.shipAssembledEncoder == nil {
		d.shipAssembledEncoder = encoder.NewShipAssembled()
	}

	return d.shipAssembledEncoder
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

func (d *diContainer) InitOutboxDispatcher() error {
	if d.outboxDispatcher != nil {
		return nil
	}
	if d.shipAssembledProducer == nil {
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
		d.shipAssembledProducer,
		dispatcherConfig,
	)
	if err != nil {
		return err
	}

	d.outboxDispatcher = dispatcher
	return nil
}

func (d *diContainer) OutboxDispatcher() *outboxdispatcher.Dispatcher {
	return d.outboxDispatcher
}

func (d *diContainer) Close(ctx context.Context) {
	if d.orderPaidConsumer != nil {
		if err := d.orderPaidConsumer.Close(ctx); err != nil {
			logger.Warn(ctx, "failed to close kafka consumer", zap.Error(err))
		}
		d.orderPaidConsumer = nil
	}

	if d.shipAssembledProducer != nil {
		if err := d.shipAssembledProducer.Close(ctx); err != nil {
			logger.Warn(ctx, "failed to close kafka producer", zap.Error(err))
		}
		d.shipAssembledProducer = nil
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
