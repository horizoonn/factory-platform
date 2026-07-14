package producer

import (
	"context"
	"errors"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
)

var ErrTopicRequired = errors.New("kafka topic is required")

type Producer interface {
	Publish(ctx context.Context, topic string, msg kafka.Message) error
}
