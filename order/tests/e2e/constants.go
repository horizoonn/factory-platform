//go:build e2e

package e2e

import "time"

const (
	projectName = "order-e2e"

	kafkaName           = "order-e2e-kafka"
	kafkaAlias          = "kafka-order-e2e"
	kafkaImage          = "confluentinc/cp-kafka:8.3.0"
	kafkaBrokerPort     = "29092"
	kafkaControllerPort = "29093"
	kafkaClusterID      = "Mk3OEYBSD34fcwNTJENDM2Qk"

	orderPaidTopic     = "order.paid.v1"
	shipAssembledTopic = "assembly.ship-assembled.v1"

	inventoryPostgresAlias    = "postgres-inventory-order-e2e"
	inventoryPostgresDatabase = "inventory"
	inventoryPostgresUsername = "inventory"
	inventoryPostgresPassword = "inventory"

	orderPostgresAlias    = "postgres-order-e2e"
	orderPostgresDatabase = "order"
	orderPostgresUsername = "order"
	orderPostgresPassword = "order"

	assemblyPostgresAlias    = "postgres-assembly-order-e2e"
	assemblyPostgresDatabase = "assembly"
	assemblyPostgresUsername = "assembly"
	assemblyPostgresPassword = "assembly"

	inventoryAppName    = "order-e2e-inventory-app"
	inventoryDockerfile = "deploy/docker/inventory/Dockerfile"
	inventoryGRPCPort   = "50051"

	paymentAppName    = "order-e2e-payment-app"
	paymentDockerfile = "deploy/docker/payment/Dockerfile"
	paymentGRPCPort   = "50052"

	assemblyAppName    = "order-e2e-assembly-app"
	assemblyDockerfile = "deploy/docker/assembly/Dockerfile"

	orderAppName    = "order-e2e-order-app"
	orderDockerfile = "deploy/docker/order/Dockerfile"
	orderHTTPPort   = "8080"

	setupTimeout    = 15 * time.Minute
	startupTimeout  = 5 * time.Minute
	teardownTimeout = time.Minute
	opTimeout       = 5 * time.Second
)
