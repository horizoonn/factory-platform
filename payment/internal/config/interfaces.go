package config

import "time"

type Config interface {
	PaymentGRPC() PaymentGRPCConfig
	App() AppConfig
}

type PaymentGRPCConfig interface {
	Address() string
}

type AppConfig interface {
	ShutdownTimeout() time.Duration
}
