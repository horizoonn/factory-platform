package outbox

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/horizoonn/factory-platform/assembly/internal/domain"
	eventsv1 "github.com/horizoonn/factory-platform/shared/pkg/proto/events/v1"
)

func TestShipAssembledWriterEnqueue(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{created: true}
	writer := NewShipAssembledWriter(repository)
	sourceEventID := uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f986001")
	event := domain.ShipAssembledEvent{
		ID:           uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f986002"),
		OrderID:      uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f986003"),
		UserID:       uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f986004"),
		BuildTimeSec: 10,
		OccurredAt:   time.Date(2026, time.July, 11, 9, 0, 10, 0, time.UTC),
	}

	created, err := writer.EnqueueShipAssembled(context.Background(), sourceEventID, event)
	if err != nil {
		t.Fatalf("EnqueueShipAssembled() error = %v", err)
	}
	if !created {
		t.Fatal("EnqueueShipAssembled() created = false, want true")
	}

	got := repository.event
	if got.ID != event.ID || got.SourceEventID != sourceEventID || got.AggregateID != event.OrderID {
		t.Fatalf("outbox identifiers = %s/%s/%s", got.ID, got.SourceEventID, got.AggregateID)
	}
	if got.Topic != shipAssembledTopic {
		t.Errorf("topic = %q, want %q", got.Topic, shipAssembledTopic)
	}
	if string(got.Key) != event.OrderID.String() {
		t.Errorf("key = %q, want %q", got.Key, event.OrderID)
	}
	if !got.AvailableAt.Equal(event.OccurredAt) {
		t.Errorf("available_at = %s, want %s", got.AvailableAt, event.OccurredAt)
	}

	var payload eventsv1.ShipAssembled
	if err := proto.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("unmarshal outbox payload: %v", err)
	}
	if payload.GetEventUuid() != event.ID.String() || payload.GetOrderUuid() != event.OrderID.String() {
		t.Errorf("payload identifiers = %q/%q", payload.GetEventUuid(), payload.GetOrderUuid())
	}
	if payload.GetUserUuid() != event.UserID.String() || payload.GetBuildTimeSec() != event.BuildTimeSec {
		t.Errorf("payload data = %q/%d", payload.GetUserUuid(), payload.GetBuildTimeSec())
	}
	if !payload.GetOccurredAt().AsTime().Equal(event.OccurredAt) {
		t.Errorf("payload occurred_at = %s, want %s", payload.GetOccurredAt().AsTime(), event.OccurredAt)
	}
}

type repositoryStub struct {
	event   Event
	created bool
	err     error
}

func (r *repositoryStub) Enqueue(_ context.Context, event Event) (bool, error) {
	r.event = event

	return r.created, r.err
}
