package shipassembled

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/horizoonn/factory-platform/platform/pkg/kafka"
	eventsv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/events/v1"
)

func TestRecordToShipAssembledEvent(t *testing.T) {
	eventID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()
	occurredAt := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)

	payload, err := proto.Marshal(&eventsv1.ShipAssembled{
		EventUuid:    eventID.String(),
		OrderUuid:    orderID.String(),
		UserUuid:     userID.String(),
		BuildTimeSec: 42,
		OccurredAt:   timestamppb.New(occurredAt),
	})
	require.NoError(t, err)

	event, err := recordToShipAssembledEvent(kafka.Record{Message: kafka.Message{Value: payload}})
	require.NoError(t, err)
	assert.Equal(t, eventID, event.ID)
	assert.Equal(t, orderID, event.OrderID)
	assert.Equal(t, userID, event.UserID)
	assert.EqualValues(t, 42, event.BuildTimeSec)
	assert.Equal(t, occurredAt, event.OccurredAt)
}

func TestRecordToShipAssembledEvent_Validation(t *testing.T) {
	validMessage := func() *eventsv1.ShipAssembled {
		return &eventsv1.ShipAssembled{
			EventUuid:    uuid.NewString(),
			OrderUuid:    uuid.NewString(),
			UserUuid:     uuid.NewString(),
			BuildTimeSec: 1,
			OccurredAt:   timestamppb.Now(),
		}
	}

	tests := []struct {
		name    string
		payload []byte
		mutate  func(*eventsv1.ShipAssembled)
	}{
		{name: "malformed protobuf", payload: []byte{0xff}},
		{name: "missing event uuid", mutate: func(msg *eventsv1.ShipAssembled) { msg.EventUuid = "" }},
		{name: "nil order uuid", mutate: func(msg *eventsv1.ShipAssembled) { msg.OrderUuid = uuid.Nil.String() }},
		{name: "invalid user uuid", mutate: func(msg *eventsv1.ShipAssembled) { msg.UserUuid = "not-a-uuid" }},
		{name: "negative build time", mutate: func(msg *eventsv1.ShipAssembled) { msg.BuildTimeSec = -1 }},
		{name: "missing occurred at", mutate: func(msg *eventsv1.ShipAssembled) { msg.OccurredAt = nil }},
		{name: "invalid occurred at", mutate: func(msg *eventsv1.ShipAssembled) {
			msg.OccurredAt = &timestamppb.Timestamp{Seconds: 253402300800}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := tt.payload
			if tt.mutate != nil {
				msg := validMessage()
				tt.mutate(msg)
				var err error
				payload, err = proto.Marshal(msg)
				require.NoError(t, err)
			}

			_, err := recordToShipAssembledEvent(kafka.Record{Message: kafka.Message{Value: payload}})
			require.Error(t, err)
		})
	}
}
