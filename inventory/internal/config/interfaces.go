package config

import "time"

type Config interface {
	InventoryGRPC() InventoryGRPCConfig
	App() AppConfig
}

type InventoryGRPCConfig interface {
	Address() string
}

type AppConfig interface {
	ShutdownTimeout() time.Duration
}
