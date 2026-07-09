//go:build e2e

package e2e

import "time"

const (
	projectName = "order-e2e"

	inventoryPostgresAlias    = "postgres-inventory-order-e2e"
	inventoryPostgresDatabase = "inventory"
	inventoryPostgresUsername = "inventory"
	inventoryPostgresPassword = "inventory"

	orderPostgresAlias    = "postgres-order-e2e"
	orderPostgresDatabase = "order"
	orderPostgresUsername = "order"
	orderPostgresPassword = "order"

	inventoryAppName    = "order-e2e-inventory-app"
	inventoryDockerfile = "deploy/docker/inventory/Dockerfile"
	inventoryGRPCPort   = "50051"

	paymentAppName    = "order-e2e-payment-app"
	paymentDockerfile = "deploy/docker/payment/Dockerfile"
	paymentGRPCPort   = "50052"

	orderAppName    = "order-e2e-order-app"
	orderDockerfile = "deploy/docker/order/Dockerfile"
	orderHTTPPort   = "8080"

	setupTimeout    = 15 * time.Minute
	startupTimeout  = 5 * time.Minute
	teardownTimeout = time.Minute
	opTimeout       = 5 * time.Second
)
