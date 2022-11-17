package eventstore

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
)

// AppendOnly appends an event into an event store
type AppendOnly interface {
	// AppendEvent pushes an event to the event store and assumes that the event has already been validated before receiving
	Append(ctx context.Context, event *event.Event) error
}
