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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	evryspb "github.com/z5labs/evrys/proto"

	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	Use:     "events -|FILE",
	Aliases: []string{"event"},
	Short:   "Publish events to evrys",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := openSource(cmd, args[0])
		if err != nil {
			return Error{
				Cmd:   cmd,
				Cause: err,
			}
		}
		zap.L().Debug("opened source", zap.String("name", args[0]))

		e := getEndpoint()
		evrys, err := dialEvrys(cmd.Context(), e)
		if err != nil {
			zap.L().Error("failed to dial evrys", zap.Error(err))
			return Error{
				Cmd:   cmd,
				Cause: err,
			}
		}
		zap.L().Debug("dialed evrys", zap.String("addr", e.Addr))

		eventCh := make(chan event.Event)

		g, gctx := errgroup.WithContext(cmd.Context())
		g.Go(readEvents(gctx, r, eventCh))
		g.Go(recordEvents(gctx, evrys, eventCh))
		return g.Wait()
	},
}

func init() {
	publishCmd.AddCommand(publishEventsCmd)

	publishEventsCmd.Flags().String("grpc-endpoint", "", "gRPC endpoint of evrys service")

	viper.BindPFlag("grpc-endpoint", publishEventsCmd.Flags().Lookup("grpc-endpoint"))
	viper.BindPFlag("format", publishEventsCmd.Flags().Lookup("format"))
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

var unknownEvrysTargetErr = errors.New("unknown evrys target type")

func dialEvrys(ctx context.Context, e endpoint) (evryspb.EvrysClient, error) {
	if e.Type != grpcEndpoint {
		return nil, UnableToDialError{
			Target: e.Addr,
			Reason: unknownEvrysTargetErr,
		}
	}

	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cc, err := grpc.DialContext(
		dialCtx,
		e.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, UnableToDialError{Target: e.Addr, Reason: err}
	}
	return evryspb.NewEvrysClient(cc), nil
}

func openSource(cmd *cobra.Command, name string) (io.Reader, error) {
	if name == "-" {
		return cmd.InOrStdin(), nil
	}
	absPath, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}
	return os.Open(absPath)
}

func readEvents(ctx context.Context, r io.Reader, eventCh chan<- event.Event) func() error {
	g, gctx := errgroup.WithContext(ctx)
	return func() error {
		defer close(eventCh)
		defer g.Wait()

		numOfEvents := 0
		br := bufio.NewReader(r)
		zap.L().Info("reading events")
		for {
			select {
			case <-gctx.Done():
			default:
			}

			line, _, err := br.ReadLine()
			if err == io.EOF {
				zap.L().Info("read events", zap.Int("num_of_events", numOfEvents))
				return nil
			}
			if err != nil {
				zap.L().Error("failed to read line", zap.Error(err))
				return err
			}

			var event event.Event
			err = event.UnmarshalJSON([]byte(line))
			if err != nil {
				zap.L().Error("failed to unmarshal event", zap.Error(err))
				return err
			}
			numOfEvents += 1

			g.Go(func() error {
				select {
				case <-gctx.Done():
				case eventCh <- event:
				}
				return nil
			})
		}
	}
}

func recordEvents(ctx context.Context, evrys evryspb.EvrysClient, eventCh <-chan event.Event) func() error {
	return func() error {
		numOfEvents := 0
		zap.L().Info("recording events")
		for {
			select {
			case <-ctx.Done():
				return nil
			case event := <-eventCh:
				if event.Type() == "" {
					zap.L().Info("recorded events", zap.Int("num_of_events", numOfEvents))
					return nil
				}
				numOfEvents += 1

				// TODO
				_, err := evrys.RecordEvent(ctx, &pb.CloudEvent{
					Id:          event.ID(),
					Source:      event.Source(),
					SpecVersion: event.SpecVersion(),
					Type:        event.Type(),
				})
				if err != nil {
					zap.L().Error("failed to record event", zap.Error(err))
					return err
				}
			}
		}
	}
}
