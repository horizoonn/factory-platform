package franz

import (
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
)

func fromKgoRecord(record *kgo.Record) kafka.Message {
	return kafka.Message{
		Topic:     record.Topic,
		Key:       record.Key,
		Value:     record.Value,
		Headers:   fromKgoHeaders(record.Headers),
		Timestamp: record.Timestamp,
		Partition: record.Partition,
		Offset:    record.Offset,
	}
}

func fromKgoHeaders(headers []kgo.RecordHeader) []kafka.Header {
	if len(headers) == 0 {
		return nil
	}

	result := make([]kafka.Header, 0, len(headers))
	for _, header := range headers {
		result = append(result, kafka.Header{
			Key:   header.Key,
			Value: header.Value,
		})
	}

	return result
}
