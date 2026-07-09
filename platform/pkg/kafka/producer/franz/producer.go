package franz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
	"github.com/horizoonn/factory-platform/platform/pkg/kafka/producer"
)

type Producer struct {
	client       *kgo.Client
	defaultTopic string
}

func NewProducer(config Config) (*Producer, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(config.Brokers...),
	}
	if config.ClientID != "" {
		opts = append(opts, kgo.ClientID(config.ClientID))
	}
	if config.DefaultTopic != "" {
		opts = append(opts, kgo.DefaultProduceTopic(config.DefaultTopic))
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer client: %w", err)
	}

	return &Producer{
		client:       client,
		defaultTopic: config.DefaultTopic,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, msg kafka.Message) error {
	if err := p.validateMessage(msg); err != nil {
		return err
	}

	record := toKgoRecord(msg)
	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("produce kafka message: %w", err)
	}

	return nil
}

func (p *Producer) PublishJSON(ctx context.Context, msg kafka.Message, payload any) error {
	value, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal kafka json payload: %w", err)
	}

	msg.Value = value

	return p.Publish(ctx, msg)
}

func (p *Producer) Send(ctx context.Context, key, value []byte) error {
	return p.Publish(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

func (p *Producer) Close() {
	if p == nil || p.client == nil {
		return
	}

	p.client.Close()
}

func (p *Producer) validateMessage(msg kafka.Message) error {
	if msg.Topic == "" && p.defaultTopic == "" {
		return producer.ErrTopicRequired
	}

	return nil
}
