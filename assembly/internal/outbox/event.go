package outbox

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID            uuid.UUID
	SourceEventID uuid.UUID
	AggregateID   uuid.UUID
	Topic         string
	Key           []byte
	Payload       []byte
	Headers       map[string]string
	AvailableAt   time.Time
	Attempts      int
}
