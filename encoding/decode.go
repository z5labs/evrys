package encoding

import (
	"bufio"
	"io"

	"github.com/cloudevents/sdk-go/v2/event"
)

// Decoder
type Decoder interface {
	Decode() (event.Event, error)
}

// JsonDecoder
type JsonDecoder struct {
	src *bufio.Reader
}

// NewJsonDecoder
func NewJsonDecoder(src io.Reader) *JsonDecoder {
	return &JsonDecoder{
		src: bufio.NewReader(src),
	}
}

// Decode
func (d *JsonDecoder) Decode() (event.Event, error) {
	line, _, err := d.src.ReadLine()
	if err == io.EOF {
		return event.Event{}, io.EOF
	}
	if err != nil {
		return event.Event{}, err
	}

	var ev event.Event
	err = ev.UnmarshalJSON(line)
	if err != nil {
		return ev, err
	}

	return ev, nil
}
