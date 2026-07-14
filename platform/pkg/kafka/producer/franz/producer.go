package franz

import (
	"context"
	"fmt"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka/producer"
)

type Producer struct {
	client *kgo.Client
}

func NewProducer(config Config) (*Producer, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(config.Brokers...),
		kgo.RecordDeliveryTimeout(config.deliveryTimeout()),
		kgo.AllowIdempotentProduceCancellation(),
	}
	if config.ClientID != "" {
		opts = append(opts, kgo.ClientID(config.ClientID))
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer client: %w", err)
	}

	return &Producer{client: client}, nil
}

func (p *Producer) Ping(ctx context.Context) error {
	if err := p.client.Ping(ctx); err != nil {
		return fmt.Errorf("ping kafka producer: %w", err)
	}

	return nil
}

func (p *Producer) Publish(ctx context.Context, topic string, msg kafka.Message) error {
	if strings.TrimSpace(topic) == "" {
		return producer.ErrTopicRequired
	}

	record := toKgoRecord(topic, msg)
	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("produce kafka message: %w", err)
	}

	return nil
}

func (p *Producer) Close(ctx context.Context) error {
	if p == nil || p.client == nil {
		return nil
	}

	err := p.client.Flush(ctx)
	if err != nil {
		// The client is being discarded; abort keeps shutdown within its deadline.
		p.client.UnsafeAbortBufferedRecords()
	}
	p.client.Close()
	if err != nil {
		return fmt.Errorf("flush kafka producer: %w", err)
	}

	return nil
}
