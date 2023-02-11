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

	"github.com/z5labs/evrys/svc-event-log/eventlogpb"

	"google.golang.org/protobuf/types/known/emptypb"
)

// Serve
func Serve(ctx context.Context) error {
	return nil
}

type service struct {
	eventlogpb.UnimplementedEventLogServer
}

// Append
func (s *service) Append(ctx context.Context, req *eventlogpb.AppendRequest) (*emptypb.Empty, error) {
	return nil, nil
}

// Iterate
func (s *service) Iterate(req *eventlogpb.IterateRequest, stream eventlogpb.EventLog_IterateServer) error {
	return nil
}
