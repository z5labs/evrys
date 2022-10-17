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
	"fmt"
	"io"
	"strings"

	evryspb "github.com/z5labs/evrys/proto"

	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type UnableToDialError struct {
	Target string
	Reason error
}

func (e UnableToDialError) Error() string {
	return fmt.Sprintf("unable to dial evrys at %s: %s", e.Target, e.Reason)
}

func (e UnableToDialError) Unwrap() error {
	return e.Reason
}

type UnsupportedEventFormatError struct {
	Format string
}

func (e UnsupportedEventFormatError) Error() string {
	return "unsupported event serialization format: " + e.Format
}

type MissingEndpointError struct{}

func (e MissingEndpointError) Error() string {
	return "must provide an evrys api endpoint"
}

var publishEventsCmd = &cobra.Command{
	Use:     "events -|EVENT...",
	Aliases: []string{"event"},
	Short:   "Publish events to evrys",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := getFormat()
		if err != nil {
			return Error{
				Cmd:   cmd,
				Cause: err,
			}
		}

		events, err := readEvents(cmd.InOrStdin(), format, args...)
		if err != nil {
			return Error{
				Cmd:   cmd,
				Cause: err,
			}
		}
		zap.L().Info("read events", zap.Int("num_of_events", len(events)))

		e := getEndpoint()
		switch e.Type {
		case grpcEndpoint:
			zap.L().Info(
				"publishing events using gRPC",
				zap.String("endpoint", e.Addr),
				zap.Int("num_of_events", len(events)),
			)

			evrys, err := dialEvrys(e.Addr, grpc.WithBlock())
			if err != nil {
				return Error{
					Cmd:   cmd,
					Cause: err,
				}
			}
			return recordEvents(cmd.Context(), evrys, events...)
		default:
			return Error{
				Cmd:   cmd,
				Cause: MissingEndpointError{},
			}
		}
	},
}

func init() {
	publishCmd.AddCommand(publishEventsCmd)

	publishEventsCmd.Flags().String("grpc-endpoint", "", "gRPC endpoint of evrys service")
	publishEventsCmd.Flags().String("format", "json", "Input format of events")

	viper.BindPFlag("grpc-endpoint", publishEventsCmd.Flags().Lookup("grpc-endpoint"))
	viper.BindPFlag("format", publishEventsCmd.Flags().Lookup("format"))
}

const (
	jsonFormat  = "json"
	protoFormat = "proto"
)

func getFormat() (string, error) {
	format := strings.ToLower(viper.GetString("format"))
	switch format {
	case jsonFormat:
		return jsonFormat, nil
	case protoFormat:
		return protoFormat, nil
	default:
		return "", UnsupportedEventFormatError{Format: format}
	}
}

const (
	grpcEndpoint = "grpc"
)

type endpoint struct {
	Type string
	Addr string
}

func getEndpoint() endpoint {
	var e endpoint
	if s := strings.TrimSpace(viper.GetString("grpc-endpoint")); s != "" {
		e.Type = grpcEndpoint
		e.Addr = s
		return e
	}
	return e
}

func readEvents(r io.Reader, format string, args ...string) ([]*pb.CloudEvent, error) {
	if args[0] == "-" {
		return []*pb.CloudEvent{}, nil
	}

	return []*pb.CloudEvent{}, nil
}

func dialEvrys(target string, opts ...grpc.DialOption) (evryspb.EvrysClient, error) {
	cc, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, UnableToDialError{Target: target, Reason: err}
	}
	return evryspb.NewEvrysClient(cc), nil
}

func recordEvents(ctx context.Context, evrys evryspb.EvrysClient, events ...*pb.CloudEvent) error {
	return nil
}
