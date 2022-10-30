package encoding

import "github.com/cloudevents/sdk-go/v2/event"

type Decoder interface {
	Decode() (*event.Event, error)
}
