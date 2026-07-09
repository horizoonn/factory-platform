package consumer

import (
	"context"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
)

type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
}

type Handler interface {
	Handle(ctx context.Context, msg kafka.Message) error
}

type HandlerFunc func(ctx context.Context, msg kafka.Message) error

func (f HandlerFunc) Handle(ctx context.Context, msg kafka.Message) error {
	return f(ctx, msg)
}

type Middleware func(Handler) Handler
