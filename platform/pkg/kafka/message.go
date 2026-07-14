package kafka

import (
	"time"
)

type Header struct {
	Key   string
	Value []byte
}

type Message struct {
	Key     []byte
	Value   []byte
	Headers []Header
}

type Record struct {
	Message

	Topic     string
	Timestamp time.Time
	Partition int32
	Offset    int64
}

func TextHeaders(headers map[string]string) []Header {
	if len(headers) == 0 {
		return nil
	}

	result := make([]Header, 0, len(headers))
	for key, value := range headers {
		result = append(result, Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	return result
}
