//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"

	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	"github.com/horizoonn/factory-platform/platform/pkg/testcontainers/app"
	tcnetwork "github.com/horizoonn/factory-platform/platform/pkg/testcontainers/network"
	"github.com/horizoonn/factory-platform/platform/pkg/testcontainers/path"
	tcpostgres "github.com/horizoonn/factory-platform/platform/pkg/testcontainers/postgres"
)

type TestEnvironment struct {
	Network  *tcnetwork.Network
	Postgres *tcpostgres.Container
	App      *app.Container
	Pool     *pgxpool.Pool
}

func setupTestEnvironment(ctx context.Context) (*TestEnvironment, error) {
	generatedNetwork, err := tcnetwork.NewNetwork(ctx, projectName)
	if err != nil {
		return nil, err
	}

	generatedPostgres, err := tcpostgres.NewContainer(ctx,
		tcpostgres.WithDatabase(postgresDatabase),
		tcpostgres.WithUsername(postgresUsername),
		tcpostgres.WithPassword(postgresPassword),
		tcpostgres.WithContainerCustomizers(
			network.WithNetworkName([]string{postgresAlias}, generatedNetwork.Name()),
		),
	)
	if err != nil {
		return nil, cleanupOnSetupError(&TestEnvironment{Network: generatedNetwork}, err)
	}

	poolConfig, err := generatedPostgres.PgxPoolConfig(ctx)
	if err != nil {
		return nil, cleanupOnSetupError(&TestEnvironment{
			Network:  generatedNetwork,
			Postgres: generatedPostgres,
		}, err)
	}

	pool, err := pgxpool.NewPool(ctx, poolConfig)
	if err != nil {
		return nil, cleanupOnSetupError(&TestEnvironment{
			Network:  generatedNetwork,
			Postgres: generatedPostgres,
		}, err)
	}

	projectRoot, err := path.ProjectRoot()
	if err != nil {
		return nil, cleanupOnSetupError(&TestEnvironment{
			Network:  generatedNetwork,
			Postgres: generatedPostgres,
			Pool:     pool,
		}, err)
	}

	appContainer, err := app.NewContainer(ctx,
		app.WithName(inventoryAppName),
		app.WithPort(inventoryGRPCPort),
		app.WithDockerfile(projectRoot, inventoryDockerfile),
		app.WithNetwork(generatedNetwork.Name()),
		app.WithLogOutput(os.Stdout),
		app.WithKeepImage(true),
		app.WithStartupWait(wait.ForExec([]string{
			"/bin/grpc-health-probe",
			"-addr=:50051",
		}).WithStartupTimeout(startupTimeout)),
		app.WithEnv(map[string]string{
			"POSTGRES_HOST":     postgresAlias,
			"POSTGRES_PORT":     "5432",
			"POSTGRES_USER":     postgresUsername,
			"POSTGRES_PASSWORD": postgresPassword,
			"POSTGRES_DB":       postgresDatabase,
			"POSTGRES_SSL_MODE": "disable",
			"POSTGRES_TIMEOUT":  "5s",
			"LOGGER_LEVEL":      "debug",
			"LOGGER_AS_JSON":    "false",
		}),
	)
	if err != nil {
		return nil, cleanupOnSetupError(&TestEnvironment{
			Network:  generatedNetwork,
			Postgres: generatedPostgres,
			Pool:     pool,
		}, err)
	}

	return &TestEnvironment{
		Network:  generatedNetwork,
		Postgres: generatedPostgres,
		App:      appContainer,
		Pool:     pool,
	}, nil
}

func cleanupOnSetupError(env *TestEnvironment, setupErr error) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), teardownTimeout)
	defer cancel()

	if err := teardownTestEnvironment(cleanupCtx, env); err != nil {
		return fmt.Errorf("%w; cleanup test environment: %v", setupErr, err)
	}

	return setupErr
}
