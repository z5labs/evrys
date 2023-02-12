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
	"net"

	"github.com/z5labs/evrys/lib/eventstore"
	"github.com/z5labs/evrys/svc-event-log/eventlogpb"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// EventStore
type EventStore interface {
	eventstore.AppendOnly
}

// ServiceConfig
type ServiceConfig struct {
	Logger     *zap.Logger
	EventStore EventStore
	Listener   net.Listener
}

// Serve
func Serve(ctx context.Context, cfg ServiceConfig) error {
	if cfg.EventStore == nil {
		return errors.New("event store must be provided")
	}
	if cfg.Listener == nil {
		return errors.New("listener must be set")
	}
	s := &service{
		log:   cfg.Logger,
		store: cfg.EventStore,
	}
	if s.log == nil {
		s.log = zap.NewNop()
	}

	grpcServer := grpc.NewServer()
	eventlogpb.RegisterEventLogServer(grpcServer, s)

	done := make(chan struct{}, 1)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		defer close(done)
		return grpcServer.Serve(cfg.Listener)
	})
	g.Go(func() error {
		select {
		case <-gctx.Done():
			grpcServer.GracefulStop()
			return gctx.Err()
		case <-done:
			return nil
		}
	})

	err := g.Wait()
	<-done
	return err
}

type service struct {
	eventlogpb.UnimplementedEventLogServer

	log   *zap.Logger
	store EventStore
}

// Append
func (s *service) Append(ctx context.Context, req *eventlogpb.AppendRequest) (*emptypb.Empty, error) {
	return nil, nil
}

// Iterate
func (s *service) Iterate(req *eventlogpb.IterateRequest, stream eventlogpb.EventLog_IterateServer) error {
	return nil
}
