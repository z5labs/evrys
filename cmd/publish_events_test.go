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

package cmd

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/z5labs/evrys/grpc"
	evryspb "github.com/z5labs/evrys/proto"

	cloudeventpb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
)

func TestPublishEvents(t *testing.T) {
	t.Run("will return an error", func(t *testing.T) {
		t.Run("if no args are given", func(t *testing.T) {
			err := Execute("publish", "events")
			if !assert.Error(t, err) {
				return
			}
			if !assert.Equal(t, "accepts 1 arg(s), received 0", err.Error()) {
				return
			}
		})

		t.Run("if event file path does not exist", func(t *testing.T) {
			err := Execute("publish", "events", "--grpc-endpoint=\"example.com:8080\"", "randomfile.json")
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, Error{}, err) {
				return
			}

			cmdErr := err.(Error)
			if !assert.Equal(t, "evrys publish events", cmdErr.Cmd.CommandPath()) {
				return
			}
			// TODO
		})

		t.Run("if an unknown source format is provided", func(t *testing.T) {
			err := Execute("publish", "events", "--source=random", "-")
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, Error{}, err) {
				return
			}

			cmdErr := err.(Error)
			if !assert.Equal(t, "evrys publish events", cmdErr.Cmd.CommandPath()) {
				return
			}
			if !assert.IsType(t, UnsupportedEventFormatError{}, cmdErr.Unwrap()) {
				return
			}
		})

		t.Run("if no evrys endpoint is provided", func(t *testing.T) {
			err := Execute("publish", "events", "-")
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, Error{}, err) {
				return
			}

			cmdErr := err.(Error)
			if !assert.Equal(t, "evrys publish events", cmdErr.Cmd.CommandPath()) {
				return
			}
			if !assert.IsType(t, MissingEndpointError{}, cmdErr.Unwrap()) {
				return
			}
		})

		t.Run("if the evrys endpoint is unreachable", func(t *testing.T) {
			err := Execute("publish", "events", "--grpc-endpoint=\"example.com:8080\"", "-")
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, Error{}, err) {
				return
			}

			cmdErr := err.(Error)
			if !assert.Equal(t, "evrys publish events", cmdErr.Cmd.CommandPath()) {
				return
			}
			if !assert.IsType(t, UnableToDialError{}, cmdErr.Unwrap()) {
				return
			}
		})
	})

	t.Run("will successfully publish events", func(t *testing.T) {
		t.Run("if the events are json encoded and newline separated", func(t *testing.T) {
			ls, err := net.Listen("tcp", ":0")
			if !assert.Nil(t, err) {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			errCh := make(chan error, 1)
			go func() {
				defer close(errCh)

				evrys := grpc.MockEvrys(
					grpc.WithRecordEvent(func(ctx context.Context, ce *cloudeventpb.CloudEvent) (*evryspb.RecordEventResponse, error) {
						return new(evryspb.RecordEventResponse), nil
					}),
				)
				err := evrys.Serve(ctx, ls)
				if err == grpc.ErrServerStopped {
					return
				}
				errCh <- err
			}()

			go func() {
				defer cancel()

				err := Execute("publish", "events", "--source=json", "--grpc-endpoint="+ls.Addr().String(), "testdata/events.json")
				if !assert.Nil(t, err) {
					return
				}
			}()

			err = <-errCh
			if !assert.Nil(t, err) {
				return
			}
		})

		t.Run("if the events are protobuf encoded and byte size prefixed", func(t *testing.T) {})
	})
}
