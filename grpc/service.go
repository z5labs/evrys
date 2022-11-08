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

	format "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cloudeventpb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// ErrServerStopped wraps grpc.ErrServerStopped
var ErrServerStopped = grpc.ErrServerStopped

var _ evryspb.EvrysServer = &EvrysService{}

// EvrysService is defines the grpc for evrys and implements the interface from evrys proto
type EvrysService struct {
	evryspb.UnimplementedEvrysServer
	eventStore EventStore
	log        *zap.Logger
}

// NewEvrysService creates an instance of EvrysService
func NewEvrysService(eventStore EventStore, logger *zap.Logger) *EvrysService {
	return &EvrysService{
		eventStore: eventStore,
		log:        logger,
	}
}

// Serve creates and runs the grpc server
func (s *EvrysService) Serve(ctx context.Context, ls net.Listener) error {
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

// GetEvent retrieves an event from the event store
func (s *EvrysService) GetEvent(ctx context.Context, req *evryspb.GetEventRequest) (*cloudeventpb.CloudEvent, error) {
	return nil, status.Error(codes.NotFound, "")
}

// RecordEvent records an event in the event store
func (s *EvrysService) RecordEvent(ctx context.Context, req *cloudeventpb.CloudEvent) (*evryspb.RecordEventResponse, error) {
	event, err := format.FromProto(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.log.Info(
		"received event",
		zap.String("event_id", event.ID()),
		zap.String("event_source", event.Source()),
		zap.String("event_type", event.Type()),
	)
	return new(evryspb.RecordEventResponse), nil
}

// SliceEvents retrieves multiple events from the event store
func (s *EvrysService) SliceEvents(req *evryspb.SliceEventsRequest, stream evryspb.Evrys_SliceEventsServer) error {
	s.log.Info("received slice request")
	return nil
}
