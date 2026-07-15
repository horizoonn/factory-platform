//go:build e2e

package e2e

import "time"

const (
	projectName = "inventory-e2e"

	postgresAlias    = "postgres-inventory"
	postgresDatabase = "inventory"
	postgresUsername = "inventory"
	postgresPassword = "inventory"

	inventoryAppName    = "inventory-app"
	inventoryDockerfile = "deploy/docker/inventory/Dockerfile"
	inventoryGRPCPort   = "50051"
	inventoryHTTPPort   = "8082"

	setupTimeout    = 15 * time.Minute
	startupTimeout  = 5 * time.Minute
	teardownTimeout = time.Minute
	opTimeout       = 5 * time.Second
)
