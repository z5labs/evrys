package grpc

import "github.com/z5labs/evrys/eventstore"

// EventStore combines single method interfaces from eventstore into one interface
type EventStore interface {
	eventstore.AppendOnly
}
