package config

import "time"

type Config interface {
	InventoryGRPC() InventoryGRPCConfig
	Migrations() MigrationsConfig
	App() AppConfig
}

type InventoryGRPCConfig interface {
	Address() string
}

type MigrationsConfig interface {
	Dir() string
}

type AppConfig interface {
	ShutdownTimeout() time.Duration
}
