package integrations

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
)

// PutEvent puts an event into an event store
type PutEvent interface {
	PutEvent(ctx context.Context, event event.Event) error
}
