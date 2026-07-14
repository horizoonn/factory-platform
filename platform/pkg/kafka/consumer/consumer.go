package consumer

import (
	"context"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
)

type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
}

type Handler interface {
	Handle(ctx context.Context, record kafka.Record) error
}

type HandlerFunc func(ctx context.Context, record kafka.Record) error

func (f HandlerFunc) Handle(ctx context.Context, record kafka.Record) error {
	return f(ctx, record)
}

type Middleware func(Handler) Handler
