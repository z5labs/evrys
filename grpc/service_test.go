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

package grpc

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	format "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	evryspb "github.com/z5labs/evrys/proto"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestEvrysService_GetEvent(t *testing.T) {
	t.Run("will return not found", func(t *testing.T) {
		t.Run("if no events exists in the store", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			ls, err := net.Listen("tcp", ":0")
			if !assert.Nil(t, err) {
				return
			}

			s := NewEvrysService(&mockEventStore{}, zap.L())

			errChan := make(chan error, 1)
			go func(ls net.Listener) {
				defer close(errChan)

				err := s.Serve(ctx, ls)
				if err != nil && err != ErrServerStopped {
					errChan <- err
				}
			}(ls)

			cc, err := grpc.Dial(ls.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if !assert.Nil(t, err) {
				return
			}

			client := evryspb.NewEvrysClient(cc)
			_, err = client.GetEvent(ctx, &evryspb.GetEventRequest{EventId: "1234"})
			if !assert.Error(t, err) {
				return
			}

			grpcStatus := status.Convert(err)
			if !assert.Equal(t, codes.NotFound, grpcStatus.Code()) {
				return
			}

			cancel()

			select {
			case <-ctx.Done():
			case err := <-errChan:
				if !assert.Nil(t, err) {
					return
				}
			}
		})
	})
}

func TestEvrysService_RecordEvent(t *testing.T) {
	t.Run("will succeed", func(t *testing.T) {
		t.Run("if the event properly follows the cloudevents spec", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			ls, err := net.Listen("tcp", ":0")
			if !assert.Nil(t, err) {
				return
			}

			s := NewEvrysService(&mockEventStore{}, zap.L())

			errChan := make(chan error, 1)
			go func(ls net.Listener) {
				defer close(errChan)

				err := s.Serve(ctx, ls)
				if err != nil && err != ErrServerStopped {
					errChan <- err
				}
			}(ls)

			cc, err := grpc.Dial(ls.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if !assert.Nil(t, err) {
				return
			}

			event, err := format.ToProto(&event.Event{})
			if !assert.Nil(t, err) {
				return
			}

			client := evryspb.NewEvrysClient(cc)
			_, err = client.RecordEvent(ctx, event)
			if !assert.Nil(t, err) {
				return
			}

			cancel()

			select {
			case <-ctx.Done():
			case err := <-errChan:
				if !assert.Nil(t, err) {
					return
				}
			}
		})
	})
}

func TestEvrysService_SliceEvents(t *testing.T) {
	t.Run("will return no events", func(t *testing.T) {
		t.Run("if the store is empty", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			ls, err := net.Listen("tcp", ":0")
			if !assert.Nil(t, err) {
				return
			}

			s := NewEvrysService(&mockEventStore{}, zap.L())

			errChan := make(chan error, 1)
			go func(ls net.Listener) {
				defer close(errChan)

				err := s.Serve(ctx, ls)
				if err != nil && err != ErrServerStopped {
					errChan <- err
				}
			}(ls)

			cc, err := grpc.Dial(ls.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if !assert.Nil(t, err) {
				return
			}

			client := evryspb.NewEvrysClient(cc)
			stream, err := client.SliceEvents(ctx, &evryspb.SliceEventsRequest{})
			if !assert.Nil(t, err) {
				return
			}

			_, err = stream.Recv()
			if !assert.Equal(t, io.EOF, err) {
				return
			}

			cancel()

			select {
			case <-ctx.Done():
			case err := <-errChan:
				if !assert.Nil(t, err) {
					return
				}
			}
		})
	})
}
