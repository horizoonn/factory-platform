//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"

	"github.com/testcontainers/testcontainers-go"
	tcnetworkcustomizer "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
	"github.com/horizoonn/factory-platform/platform/pkg/testcontainers/app"
	tcnetwork "github.com/horizoonn/factory-platform/platform/pkg/testcontainers/network"
	"github.com/horizoonn/factory-platform/platform/pkg/testcontainers/path"
	tcpostgres "github.com/horizoonn/factory-platform/platform/pkg/testcontainers/postgres"
)

type TestEnvironment struct {
	Network *tcnetwork.Network
	Kafka   testcontainers.Container

	InventoryPostgres *tcpostgres.Container
	InventoryPool     *pgxpool.Pool
	InventoryApp      *app.Container

	PaymentApp *app.Container

	AssemblyPostgres *tcpostgres.Container
	AssemblyPool     *pgxpool.Pool
	AssemblyApp      testcontainers.Container

	OrderPostgres *tcpostgres.Container
	OrderPool     *pgxpool.Pool
	OrderApp      *app.Container
}

func setupTestEnvironment(ctx context.Context) (*TestEnvironment, error) {
	generatedNetwork, err := tcnetwork.NewNetwork(ctx, projectName)
	if err != nil {
		return nil, err
	}

	env := &TestEnvironment{Network: generatedNetwork}

	projectRoot, err := path.ProjectRoot()
	if err != nil {
		return nil, cleanupOnSetupError(env, err)
	}

	if err := setupKafka(ctx, env); err != nil {
		return nil, cleanupOnSetupError(env, err)
	}

	if err := setupInventory(ctx, env, projectRoot); err != nil {
		return nil, cleanupOnSetupError(env, err)
	}

	if err := setupPayment(ctx, env, projectRoot); err != nil {
		return nil, cleanupOnSetupError(env, err)
	}

	if err := setupAssembly(ctx, env, projectRoot); err != nil {
		return nil, cleanupOnSetupError(env, err)
	}

	if err := setupOrder(ctx, env, projectRoot); err != nil {
		return nil, cleanupOnSetupError(env, err)
	}

	return env, nil
}

func setupKafka(ctx context.Context, env *TestEnvironment) error {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:     kafkaName,
			Image:    kafkaImage,
			Networks: []string{env.Network.Name()},
			NetworkAliases: map[string][]string{
				env.Network.Name(): {kafkaAlias},
			},
			Env: map[string]string{
				"KAFKA_PROCESS_ROLES":                            "controller,broker",
				"KAFKA_NODE_ID":                                  "1",
				"KAFKA_CONTROLLER_QUORUM_VOTERS":                 "1@" + kafkaAlias + ":" + kafkaControllerPort,
				"KAFKA_LISTENERS":                                "PLAINTEXT://0.0.0.0:" + kafkaBrokerPort + ",CONTROLLER://0.0.0.0:" + kafkaControllerPort,
				"KAFKA_ADVERTISED_LISTENERS":                     "PLAINTEXT://" + kafkaAlias + ":" + kafkaBrokerPort,
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT",
				"KAFKA_INTER_BROKER_LISTENER_NAME":               "PLAINTEXT",
				"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
				"KAFKA_AUTO_CREATE_TOPICS_ENABLE":                "true",
				"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
				"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
				"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
				"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
				"CLUSTER_ID":                                     kafkaClusterID,
			},
			WaitingFor: wait.ForExec([]string{
				"bash",
				"-c",
				"kafka-topics --bootstrap-server localhost:" + kafkaBrokerPort + " --list",
			}).WithStartupTimeout(startupTimeout),
		},
		Started: true,
	})
	if err != nil {
		return fmt.Errorf("start kafka: %w", err)
	}

	env.Kafka = container
	return nil
}

func setupInventory(ctx context.Context, env *TestEnvironment, projectRoot string) error {
	container, err := tcpostgres.NewContainer(
		ctx,
		tcpostgres.WithDatabase(inventoryPostgresDatabase),
		tcpostgres.WithUsername(inventoryPostgresUsername),
		tcpostgres.WithPassword(inventoryPostgresPassword),
		tcpostgres.WithContainerCustomizers(
			tcnetworkcustomizer.WithNetworkName([]string{inventoryPostgresAlias}, env.Network.Name()),
		),
	)
	if err != nil {
		return fmt.Errorf("start inventory postgres: %w", err)
	}
	env.InventoryPostgres = container

	poolConfig, err := container.PgxPoolConfig(ctx)
	if err != nil {
		return fmt.Errorf("get inventory postgres pool config: %w", err)
	}

	pool, err := pgxpool.NewPool(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("create inventory postgres pool: %w", err)
	}
	env.InventoryPool = pool

	inventoryApp, err := app.NewContainer(
		ctx,
		app.WithName(inventoryAppName),
		app.WithPort(inventoryGRPCPort),
		app.WithDockerfile(projectRoot, inventoryDockerfile),
		app.WithNetwork(env.Network.Name()),
		app.WithLogOutput(os.Stdout),
		app.WithKeepImage(true),
		app.WithStartupWait(wait.ForExec([]string{
			"/bin/grpc-health-probe",
			"-addr=:50051",
		}).WithStartupTimeout(startupTimeout)),
		app.WithEnv(map[string]string{
			"POSTGRES_HOST":     inventoryPostgresAlias,
			"POSTGRES_PORT":     "5432",
			"POSTGRES_USER":     inventoryPostgresUsername,
			"POSTGRES_PASSWORD": inventoryPostgresPassword,
			"POSTGRES_DB":       inventoryPostgresDatabase,
			"POSTGRES_SSL_MODE": "disable",
			"POSTGRES_TIMEOUT":  "5s",
			"LOGGER_LEVEL":      "debug",
			"LOGGER_AS_JSON":    "false",
		}),
	)
	if err != nil {
		return fmt.Errorf("start inventory app: %w", err)
	}
	env.InventoryApp = inventoryApp

	return nil
}

func setupPayment(ctx context.Context, env *TestEnvironment, projectRoot string) error {
	paymentApp, err := app.NewContainer(
		ctx,
		app.WithName(paymentAppName),
		app.WithPort(paymentGRPCPort),
		app.WithDockerfile(projectRoot, paymentDockerfile),
		app.WithNetwork(env.Network.Name()),
		app.WithLogOutput(os.Stdout),
		app.WithKeepImage(true),
		app.WithStartupWait(wait.ForExec([]string{
			"/bin/grpc-health-probe",
			"-addr=:50052",
		}).WithStartupTimeout(startupTimeout)),
		app.WithEnv(map[string]string{
			"PAYMENT_GRPC_HOST": "0.0.0.0",
			"PAYMENT_GRPC_PORT": paymentGRPCPort,
			"LOGGER_LEVEL":      "debug",
			"LOGGER_AS_JSON":    "false",
		}),
	)
	if err != nil {
		return fmt.Errorf("start payment app: %w", err)
	}
	env.PaymentApp = paymentApp

	return nil
}

func setupAssembly(ctx context.Context, env *TestEnvironment, projectRoot string) error {
	container, err := tcpostgres.NewContainer(
		ctx,
		tcpostgres.WithDatabase(assemblyPostgresDatabase),
		tcpostgres.WithUsername(assemblyPostgresUsername),
		tcpostgres.WithPassword(assemblyPostgresPassword),
		tcpostgres.WithContainerCustomizers(
			tcnetworkcustomizer.WithNetworkName([]string{assemblyPostgresAlias}, env.Network.Name()),
		),
	)
	if err != nil {
		return fmt.Errorf("start assembly postgres: %w", err)
	}
	env.AssemblyPostgres = container

	poolConfig, err := container.PgxPoolConfig(ctx)
	if err != nil {
		return fmt.Errorf("get assembly postgres pool config: %w", err)
	}

	pool, err := pgxpool.NewPool(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("create assembly postgres pool: %w", err)
	}
	env.AssemblyPool = pool

	assemblyApp, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name: assemblyAppName,
			FromDockerfile: testcontainers.FromDockerfile{
				Context:        projectRoot,
				Dockerfile:     assemblyDockerfile,
				BuildLogWriter: os.Stdout,
				KeepImage:      true,
			},
			Networks: []string{env.Network.Name()},
			Env: map[string]string{
				"ASSEMBLY_MIGRATIONS_DIR":   "/app/migrations",
				"POSTGRES_HOST":             assemblyPostgresAlias,
				"POSTGRES_PORT":             "5432",
				"POSTGRES_USER":             assemblyPostgresUsername,
				"POSTGRES_PASSWORD":         assemblyPostgresPassword,
				"POSTGRES_DB":               assemblyPostgresDatabase,
				"POSTGRES_SSL_MODE":         "disable",
				"POSTGRES_TIMEOUT":          "5s",
				"KAFKA_BROKERS":             kafkaBrokerAddress(),
				"ORDER_PAID_CONSUMER_TOPIC": orderPaidTopic,
				"LOGGER_LEVEL":              "debug",
				"LOGGER_AS_JSON":            "false",
				"APP_SHUTDOWN_TIMEOUT":      "10s",
			},
			WaitingFor: wait.ForLog("assembly service started").
				WithStartupTimeout(startupTimeout),
		},
		Started: true,
	})
	if err != nil {
		return fmt.Errorf("start assembly app: %w", err)
	}
	env.AssemblyApp = assemblyApp

	return nil
}

func setupOrder(ctx context.Context, env *TestEnvironment, projectRoot string) error {
	container, err := tcpostgres.NewContainer(
		ctx,
		tcpostgres.WithDatabase(orderPostgresDatabase),
		tcpostgres.WithUsername(orderPostgresUsername),
		tcpostgres.WithPassword(orderPostgresPassword),
		tcpostgres.WithContainerCustomizers(
			tcnetworkcustomizer.WithNetworkName([]string{orderPostgresAlias}, env.Network.Name()),
		),
	)
	if err != nil {
		return fmt.Errorf("start order postgres: %w", err)
	}
	env.OrderPostgres = container

	poolConfig, err := container.PgxPoolConfig(ctx)
	if err != nil {
		return fmt.Errorf("get order postgres pool config: %w", err)
	}

	pool, err := pgxpool.NewPool(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("create order postgres pool: %w", err)
	}
	env.OrderPool = pool

	orderApp, err := app.NewContainer(
		ctx,
		app.WithName(orderAppName),
		app.WithPort(orderHTTPPort),
		app.WithDockerfile(projectRoot, orderDockerfile),
		app.WithNetwork(env.Network.Name()),
		app.WithLogOutput(os.Stdout),
		app.WithKeepImage(true),
		app.WithStartupWait(wait.ForHTTP("/health").
			WithPort(orderHTTPPort+"/tcp").
			WithStartupTimeout(startupTimeout)),
		app.WithEnv(map[string]string{
			"ORDER_HTTP_HOST":               "0.0.0.0",
			"ORDER_HTTP_PORT":               orderHTTPPort,
			"ORDER_MIGRATIONS_DIR":          "/app/migrations",
			"INVENTORY_GRPC_HOST":           inventoryAppName,
			"INVENTORY_GRPC_PORT":           inventoryGRPCPort,
			"PAYMENT_GRPC_HOST":             paymentAppName,
			"PAYMENT_GRPC_PORT":             paymentGRPCPort,
			"KAFKA_BROKERS":                 kafkaBrokerAddress(),
			"SHIP_ASSEMBLED_CONSUMER_TOPIC": shipAssembledTopic,
			"POSTGRES_HOST":                 orderPostgresAlias,
			"POSTGRES_PORT":                 "5432",
			"POSTGRES_USER":                 orderPostgresUsername,
			"POSTGRES_PASSWORD":             orderPostgresPassword,
			"POSTGRES_DB":                   orderPostgresDatabase,
			"POSTGRES_SSL_MODE":             "disable",
			"POSTGRES_TIMEOUT":              "5s",
			"LOGGER_LEVEL":                  "debug",
			"LOGGER_AS_JSON":                "false",
			"APP_SHUTDOWN_TIMEOUT":          "10s",
		}),
	)
	if err != nil {
		return fmt.Errorf("start order app: %w", err)
	}
	env.OrderApp = orderApp

	return nil
}

func kafkaBrokerAddress() string {
	return kafkaAlias + ":" + kafkaBrokerPort
}

func cleanupOnSetupError(env *TestEnvironment, setupErr error) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), teardownTimeout)
	defer cancel()

	if err := teardownTestEnvironment(cleanupCtx, env); err != nil {
		return fmt.Errorf("%w; cleanup test environment: %v", setupErr, err)
	}

	return setupErr
}
