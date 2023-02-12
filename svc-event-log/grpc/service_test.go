// Copyright 2023 Z5Labs and Contributors
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
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockEventStore struct {
	append func(context.Context, *event.Event) error
}

func (s mockEventStore) Append(ctx context.Context, ev *event.Event) error {
	return s.append(ctx, ev)
}

func ExampleServe() {
	ls, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cfg := ServiceConfig{
		EventStore: mockEventStore{},
		Listener:   ls,
	}

	err = Serve(ctx, cfg)
	if err != context.DeadlineExceeded {
		fmt.Println(err)
		return
	}
	// Output:
}

type mockListener struct {
	accept func() (net.Conn, error)
	close  func() error
	addr   func() net.Addr
}

func (l mockListener) Accept() (net.Conn, error) {
	return l.accept()
}

func (l mockListener) Close() error {
	return l.close()
}

func (l mockListener) Addr() net.Addr {
	return l.addr()
}

func TestServe(t *testing.T) {
	t.Run("will return an error", func(t *testing.T) {
		t.Run("if no event store is provided", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			cfg := ServiceConfig{
				Logger:   zap.NewNop(),
				Listener: &net.TCPListener{},
			}
			err := Serve(ctx, cfg)
			if !assert.Error(t, err) {
				return
			}
		})

		t.Run("if no listener is provided", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			cfg := ServiceConfig{
				Logger:     zap.NewNop(),
				EventStore: mockEventStore{},
			}
			err := Serve(ctx, cfg)
			if !assert.Error(t, err) {
				return
			}
		})

		t.Run("if the listener fails to accept a connection", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			acceptErr := errors.New("accept failed")
			cfg := ServiceConfig{
				Logger:     zap.NewNop(),
				EventStore: mockEventStore{},
				Listener: mockListener{
					accept: func() (net.Conn, error) {
						return nil, acceptErr
					},
					close: func() error { return nil },
					addr:  func() net.Addr { return &net.IPAddr{} },
				},
			}
			err := Serve(ctx, cfg)
			if !assert.Equal(t, acceptErr, err) {
				return
			}
		})
	})
}

func TestService_Append(t *testing.T) {
	t.Run("will return an error", func(t *testing.T) {
		t.Run("if it fails to unmarshal cloudevent", func(t *testing.T) {})

		t.Run("if the cloudevent is invalid", func(t *testing.T) {})

		t.Run("if the event store implementation fails to append", func(t *testing.T) {})
	})

	t.Run("will return an empty response", func(t *testing.T) {
		t.Run("if the event is valid and the event store append operation is successful", func(t *testing.T) {

		})
	})
}
