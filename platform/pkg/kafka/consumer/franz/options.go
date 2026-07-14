package franz

import (
	"context"

	"go.uber.org/zap"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer"
	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
}

type Option func(*options)

type options struct {
	logger      Logger
	middlewares []consumer.Middleware
}

func WithLogger(logger Logger) Option {
	return func(o *options) {
		if logger == nil {
			return
		}

		o.logger = logger
	}
}

func WithMiddlewares(middlewares ...consumer.Middleware) Option {
	return func(o *options) {
		o.middlewares = append(o.middlewares, middlewares...)
	}
}

func buildOptions(opts ...Option) options {
	options := options{
		logger: logger.NewNop(),
	}

	for _, opt := range opts {
		opt(&options)
	}

	return options
}
