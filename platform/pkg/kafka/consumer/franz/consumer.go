package franz

import (
	"context"
	"errors"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka/consumer"
)

var ErrClientClosed = errors.New("kafka client closed")

type Consumer struct {
	client         *kgo.Client
	logger         Logger
	middlewares    []consumer.Middleware
	maxPollRecords int
}

func NewConsumer(config Config, opts ...Option) (*Consumer, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	options := buildOptions(opts...)

	clientOpts := []kgo.Opt{
		kgo.SeedBrokers(config.Brokers...),
		kgo.ConsumerGroup(config.GroupID),
		kgo.ConsumeTopics(config.Topics...),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
	}
	if config.ClientID != "" {
		clientOpts = append(clientOpts, kgo.ClientID(config.ClientID))
	}

	client, err := kgo.NewClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer client: %w", err)
	}

	return &Consumer{
		client:         client,
		logger:         options.logger,
		middlewares:    options.middlewares,
		maxPollRecords: config.pollLimit(),
	}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler consumer.Handler) error {
	if handler == nil {
		return errors.New("kafka message handler is required")
	}

	handler = applyMiddlewares(handler, c.middlewares)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := c.pollAndHandle(ctx, handler); err != nil {
			if errors.Is(err, ErrClientClosed) {
				c.logger.Info(ctx, "kafka consumer client closed")
				return nil
			}

			c.logger.Error(ctx, "kafka consume error", zap.Error(err))
			return err
		}
	}
}

func (c *Consumer) Close() {
	if c == nil || c.client == nil {
		return
	}

	c.client.CloseAllowingRebalance()
}

func (c *Consumer) pollAndHandle(ctx context.Context, handler consumer.Handler) error {
	fetches := c.client.PollRecords(ctx, c.maxPollRecords)
	defer c.client.AllowRebalance()

	if fetches.IsClientClosed() {
		return ErrClientClosed
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := firstFetchError(fetches); err != nil {
		return err
	}

	processed := make([]*kgo.Record, 0, c.maxPollRecords)
	var handleErr error

	fetches.EachRecord(func(record *kgo.Record) {
		if handleErr != nil {
			return
		}

		msg := fromKgoRecord(record)
		if err := handler.Handle(ctx, msg); err != nil {
			handleErr = fmt.Errorf(
				"handle kafka message topic=%s partition=%d offset=%d: %w",
				record.Topic,
				record.Partition,
				record.Offset,
				err,
			)

			return
		}

		processed = append(processed, record)
	})

	if len(processed) > 0 {
		if err := c.client.CommitRecords(ctx, processed...); err != nil {
			return fmt.Errorf("commit kafka records: %w", err)
		}
	}

	return handleErr
}

func firstFetchError(fetches kgo.Fetches) error {
	for _, fetchErr := range fetches.Errors() {
		return fmt.Errorf(
			"fetch kafka records topic=%s partition=%d: %w",
			fetchErr.Topic,
			fetchErr.Partition,
			fetchErr.Err,
		)
	}

	return nil
}

func applyMiddlewares(handler consumer.Handler, middlewares []consumer.Middleware) consumer.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}
