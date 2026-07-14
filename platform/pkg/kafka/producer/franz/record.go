package franz

import (
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
)

func toKgoRecord(topic string, msg kafka.Message) *kgo.Record {
	return &kgo.Record{
		Topic:   topic,
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: toKgoHeaders(msg.Headers),
	}
}

func toKgoHeaders(headers []kafka.Header) []kgo.RecordHeader {
	if len(headers) == 0 {
		return nil
	}

	result := make([]kgo.RecordHeader, 0, len(headers))
	for _, header := range headers {
		result = append(result, kgo.RecordHeader{
			Key:   header.Key,
			Value: header.Value,
		})
	}

	return result
}
