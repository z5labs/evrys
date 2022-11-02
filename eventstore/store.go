package eventstore

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
)

// PutEvent puts an event into an event store
type PutEvent interface {
	// PutEvent pushes an event to the event store and assumes that the event has already been validated before recieving
	PutEvent(ctx context.Context, event *event.Event) error
}
