package postgres

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"

	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

const (
	DefaultImage    = "postgres:16-alpine"
	DefaultDatabase = "test"
	DefaultUsername = "test"
	DefaultPassword = "test"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
}

type Config struct {
	Image    string
	Database string
	Username string
	Password string
	Logger   Logger

	ContainerCustomizers []testcontainers.ContainerCustomizer
}

func buildConfig(opts ...Option) Config {
	cfg := Config{
		Image:    DefaultImage,
		Database: DefaultDatabase,
		Username: DefaultUsername,
		Password: DefaultPassword,
		Logger:   logger.NewNop(),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}
