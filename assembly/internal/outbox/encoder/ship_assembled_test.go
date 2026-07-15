package encoder

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
	eventsv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/events/v1"
)

func TestShipAssembled_Encode(t *testing.T) {
	const topic = "assembly.ship-assembled.v1"

	event := domain.ShipAssembledEvent{
		ID:           uuid.New(),
		OrderID:      uuid.New(),
		UserID:       uuid.New(),
		BuildTimeSec: 10,
		OccurredAt:   time.Date(2026, time.July, 14, 12, 0, 0, 0, time.UTC),
	}

	encoded, err := NewShipAssembled(topic).Encode(event)
	require.NoError(t, err)
	assert.Equal(t, event.ID, encoded.ID)
	assert.Equal(t, event.OrderID, encoded.AggregateID)
	assert.Equal(t, topic, encoded.Topic)
	assert.Equal(t, []byte(event.OrderID.String()), encoded.Key)
	assert.Equal(t, map[string]string{"event-type": shipAssembledType}, encoded.Headers)
	assert.Equal(t, event.OccurredAt, encoded.AvailableAt)

	var message eventsv1.ShipAssembled
	require.NoError(t, proto.Unmarshal(encoded.Payload, &message))
	assert.Equal(t, event.ID.String(), message.GetEventUuid())
	assert.Equal(t, event.OrderID.String(), message.GetOrderUuid())
	assert.Equal(t, event.UserID.String(), message.GetUserUuid())
	assert.Equal(t, event.BuildTimeSec, message.GetBuildTimeSec())
	require.NotNil(t, message.GetOccurredAt())
	assert.Equal(t, event.OccurredAt, message.GetOccurredAt().AsTime())
}

func TestShipAssembled_Encode_InvalidTimestamp(t *testing.T) {
	event := domain.ShipAssembledEvent{
		ID:           uuid.New(),
		OrderID:      uuid.New(),
		UserID:       uuid.New(),
		BuildTimeSec: 10,
		OccurredAt:   time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, err := NewShipAssembled("assembly.ship-assembled.v1").Encode(event)
	require.Error(t, err)
}
