package kafka

import (
	"context"
	"time"

	"go.uber.org/zap"

	platformkafka "github.com/horizoonn/factory-platform/platform/pkg/kafka"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
}

func Logging(logger Logger) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, record platformkafka.Record) error {
			startedAt := time.Now()

			if err := next.Handle(ctx, record); err != nil {
				return err
			}

			logger.Info(
				ctx,
				"kafka record handled",
				zap.String("topic", record.Topic),
				zap.Int32("partition", record.Partition),
				zap.Int64("offset", record.Offset),
				zap.Duration("duration", time.Since(startedAt)),
			)

			return nil
		})
	}
}
