package app

import (
	"context"

	inventoryapi "github.com/horizoonn/factory-platform/inventory/internal/api/inventory/v1"
	"github.com/horizoonn/factory-platform/inventory/internal/config"
	partrepo "github.com/horizoonn/factory-platform/inventory/internal/repository/part"
	partservice "github.com/horizoonn/factory-platform/inventory/internal/service/part"
	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

type diContainer struct {
	cfg config.Config

	inventoryV1API   inventorypb.InventoryServiceServer
	inventoryService inventoryapi.InventoryService
	inventoryRepo    partservice.Repository

	postgresPool *pgxpool.Pool
}

func newDIContainer(cfg config.Config) *diContainer {
	return &diContainer{
		cfg: cfg,
	}
}

func (d *diContainer) InventoryV1API() inventorypb.InventoryServiceServer {
	if d.inventoryV1API == nil {
		d.inventoryV1API = inventoryapi.NewServer(d.InventoryService())
	}

	return d.inventoryV1API
}

func (d *diContainer) InventoryService() inventoryapi.InventoryService {
	if d.inventoryService == nil {
		d.inventoryService = partservice.NewService(d.PartRepository())
	}

	return d.inventoryService
}

func (d *diContainer) PartRepository() partservice.Repository {
	if d.inventoryRepo == nil {
		if d.postgresPool == nil {
			panic("postgres pool is not initialized")
		}
		d.inventoryRepo = partrepo.NewRepository(d.postgresPool)
	}

	return d.inventoryRepo
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

func (d *diContainer) PostgresPool() *pgxpool.Pool {
	return d.postgresPool
}

func (d *diContainer) Close(_ context.Context) {
	if d.postgresPool != nil {
		d.postgresPool.Close()
		d.postgresPool = nil
	}
}
