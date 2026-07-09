package producer

import (
	"context"
	"errors"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
)

var ErrTopicRequired = errors.New("kafka topic is required")

type Producer interface {
	Publish(ctx context.Context, msg kafka.Message) error
	PublishJSON(ctx context.Context, msg kafka.Message, payload any) error
	Send(ctx context.Context, key, value []byte) error
}
