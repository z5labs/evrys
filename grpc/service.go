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
	"fmt"
	"net"

	"github.com/go-playground/validator/v10"
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

// EvrysServiceConfig stores configuration values for evrys service
type EvrysServiceConfig struct {
	EventStore EventStore  `validate:"required"`
	Log        *zap.Logger `validate:"required"`
}

// EvrysService is defines the grpc for evrys and implements the interface from evrys proto
type EvrysService struct {
	evryspb.UnimplementedEvrysServer
	config EvrysServiceConfig
}

// NewEvrysService creates an instance of EvrysService
func NewEvrysService(config EvrysServiceConfig) (*EvrysService, error) {
	if err := validator.New().Struct(&config); err != nil {
		return nil, fmt.Errorf("failed to validate config")
	}

	return &EvrysService{
		config: config,
	}, nil
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
	s.config.Log.Info(
		"received event",
		zap.String("event_id", event.ID()),
		zap.String("event_source", event.Source()),
		zap.String("event_type", event.Type()),
	)

	err = event.Validate()
	if err != nil {
		s.config.Log.Error(
			"cloud event failed to validate",
			zap.Error(err),
			zap.String("event_id", event.ID()),
			zap.String("event_source", event.Source()),
			zap.String("event_type", event.Type()),
		)
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("cloud event failed validation: %s", err.Error()))
	}

	err = s.config.EventStore.Append(ctx, event)
	if err != nil {
		s.config.Log.Error(
			"failed to append event to event store",
			zap.Error(err),
			zap.String("event_id", event.ID()),
			zap.String("event_source", event.Source()),
			zap.String("event_type", event.Type()),
		)

		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to append event to event store: %s", err.Error()))
	}

	return new(evryspb.RecordEventResponse), nil
}

// SliceEvents retrieves multiple events from the event store
func (s *EvrysService) SliceEvents(req *evryspb.SliceEventsRequest, stream evryspb.Evrys_SliceEventsServer) error {
	s.config.Log.Info("received slice request")
	return nil
}
