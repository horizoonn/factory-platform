package config

import "time"

type Config interface {
	OrderHTTP() OrderHTTPConfig
	InventoryGRPC() InventoryGRPCConfig
	PaymentGRPC() PaymentGRPCConfig
	Migrations() MigrationsConfig
	App() AppConfig
}

type OrderHTTPConfig interface {
	Address() string
}

type InventoryGRPCConfig interface {
	Address() string
}

type PaymentGRPCConfig interface {
	Address() string
}

type MigrationsConfig interface {
	Dir() string
}

type AppConfig interface {
	ShutdownTimeout() time.Duration
}
