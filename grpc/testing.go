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
	"net"

	evryspb "github.com/z5labs/evrys/proto"

	cloudeventpb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/cloudevents/sdk-go/v2/event"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type mockEventStore struct {
}

func (m *mockEventStore) PutEvent(ctx context.Context, event *event.Event) error {
	return nil
}

func MockEvrys(opts ...func(MockEvrysService) MockEvrysService) MockEvrysService {
	s := MockEvrysService{}
	for _, opt := range opts {
		s = opt(s)
	}
	return s
}

func WithGetEvent(f func(context.Context, *evryspb.GetEventRequest) (*cloudeventpb.CloudEvent, error)) func(MockEvrysService) MockEvrysService {
	return func(mes MockEvrysService) MockEvrysService {
		mes.eventGetter = f
		return mes
	}
}

func WithRecordEvent(f func(context.Context, *cloudeventpb.CloudEvent) (*evryspb.RecordEventResponse, error)) func(MockEvrysService) MockEvrysService {
	return func(mes MockEvrysService) MockEvrysService {
		mes.eventRecorder = f
		return mes
	}
}

func WithSliceEvents(f func(*evryspb.SliceEventsRequest, evryspb.Evrys_SliceEventsServer) error) func(MockEvrysService) MockEvrysService {
	return func(mes MockEvrysService) MockEvrysService {
		mes.eventSlicer = f
		return mes
	}
}

type MockEvrysService struct {
	evryspb.UnimplementedEvrysServer

	eventGetter   func(context.Context, *evryspb.GetEventRequest) (*cloudeventpb.CloudEvent, error)
	eventRecorder func(context.Context, *cloudeventpb.CloudEvent) (*evryspb.RecordEventResponse, error)
	eventSlicer   func(*evryspb.SliceEventsRequest, evryspb.Evrys_SliceEventsServer) error
}

func (s MockEvrysService) Serve(ctx context.Context, ls net.Listener) error {
	grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	evryspb.RegisterEvrysServer(grpcServer, s)

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		err := grpcServer.Serve(ls)
		errCh <- err
	}()

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	select {
	case <-cctx.Done():
		grpcServer.GracefulStop()
		<-errCh
		return ErrServerStopped
	case err := <-errCh:
		cancel()
		return err
	}
}

func (s MockEvrysService) GetEvent(ctx context.Context, req *evryspb.GetEventRequest) (*cloudeventpb.CloudEvent, error) {
	return s.eventGetter(ctx, req)
}

func (s MockEvrysService) RecordEvent(ctx context.Context, event *cloudeventpb.CloudEvent) (*evryspb.RecordEventResponse, error) {
	return s.eventRecorder(ctx, event)
}

func (s MockEvrysService) SliceEvents(req *evryspb.SliceEventsRequest, stream evryspb.Evrys_SliceEventsServer) error {
	return s.eventSlicer(req, stream)
}
