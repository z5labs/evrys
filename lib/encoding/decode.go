// Copyright 2022 Z5Labs and Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
