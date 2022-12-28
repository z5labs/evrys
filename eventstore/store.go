package eventstore

import (
	"context"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
)

// AppendOnly appends an event into an event store
type AppendOnly interface {
	// AppendEvent pushes an event to the event store and assumes that the event has already been validated before receiving
	Append(ctx context.Context, event *event.Event) error
}

// Query looks up events in event store
type Query interface {
	GetByID(ctx context.Context, id string) (*event.Event, error)
	EventsBefore(ctx context.Context, until time.Time) (events []*event.Event, cursor any, err error)
}
