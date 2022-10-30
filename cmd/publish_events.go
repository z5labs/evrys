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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/z5labs/evrys/encoding"
	evryspb "github.com/z5labs/evrys/proto"

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

func withPublishEventsCmd() func(*viper.Viper) *cobra.Command {
	return func(v *viper.Viper) *cobra.Command {
		cmd := &cobra.Command{
			Use:     "events -|FILE...",
			Aliases: []string{"event"},
			Short:   "Publish events to evrys",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				src := cmd.InOrStdin()
				filename := strings.TrimSpace(args[0])
				if filename != "-" {
					f, err := openFile(filename)
					if err != nil {
						return Error{
							Cmd:   cmd,
							Cause: err,
						}
					}
					src = f
				}

				dec, err := getDecoder(src, v.GetString("source"))
				if err != nil {
					return Error{
						Cmd:   cmd,
						Cause: err,
					}
				}

				e := getEndpoint(v)
				evrys, err := dialEvrys(cmd.Context(), e)
				if err != nil {
					zap.L().Error("failed to dial evrys", zap.Error(err))
					return Error{
						Cmd:   cmd,
						Cause: err,
					}
				}

				eventCh := make(chan *event.Event)
				g1, g1ctx := errgroup.WithContext(cmd.Context())
				g1.Go(readEvents(g1ctx, dec, eventCh))
				g1.Go(publishEvents(g1ctx, eventCh, evrys))

				g2, g2ctx := errgroup.WithContext(g1ctx)
				g2.Go(g1.Wait)
				g2.Go(func() error {
					for {
						select {
						case <-g1ctx.Done():
							return nil
						case <-g2ctx.Done():
							return nil
						case <-time.After(1 * time.Second):
						}

						var memStats runtime.MemStats
						runtime.ReadMemStats(&memStats)

						zap.L().Debug(
							"runtime status",
							zap.Int("num_of_goroutines", runtime.NumGoroutine()),
							zap.Uint64("heap_alloc_bytes", memStats.HeapAlloc),
						)
					}
				})

				return g2.Wait()
			},
		}

		cmd.Flags().String("grpc-endpoint", "", "gRPC endpoint of evrys service")
		cmd.Flags().String("source", "json", "Source format of events")

		v.BindPFlag("grpc-endpoint", cmd.Flags().Lookup("grpc-endpoint"))
		v.BindPFlag("source", cmd.Flags().Lookup("source"))

		return cmd
	}
}

func getDecoder(r io.Reader, format string) (encoding.Decoder, error) {
	format = strings.TrimSpace(format)
	format = strings.ToLower(format)
	fmt.Println("hello", format)
	switch format {
	case "json":
		return nil, nil // TODO
	case "proto":
		return nil, nil // TODO
	default:
		return nil, UnsupportedEventFormatError{Format: format}
	}
}

func openFile(filename string) (*os.File, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	return os.Open(absPath)
}

const (
	grpcEndpoint = "grpc"
)

type endpoint struct {
	Type string
	Addr string
}

func getEndpoint(v *viper.Viper) endpoint {
	var e endpoint
	if s := strings.TrimSpace(v.GetString("grpc-endpoint")); s != "" {
		e.Type = grpcEndpoint
		e.Addr = s
		return e
	}
	return e
}

func dialEvrys(ctx context.Context, e endpoint) (evryspb.EvrysClient, error) {
	if e.Type != "grpc" {
		return nil, MissingEndpointError{}
	}

	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cc, err := grpc.DialContext(
		dialCtx,
		e.Addr,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, UnableToDialError{Target: e.Addr, Reason: err}
	}
	return evryspb.NewEvrysClient(cc), nil
}

func readEvents(ctx context.Context, dec encoding.Decoder, eventCh chan<- *event.Event) func() error {
	g, gctx := errgroup.WithContext(ctx)
	return func() error {
		defer close(eventCh)
		defer g.Wait()

		i := 0
		zap.L().Info("decoding cloudevents")
		for {
			select {
			case <-gctx.Done():
				zap.L().Warn("context cancelled during decoding of events", zap.Int("num_of_events", i))
				return nil
			default:
			}

			zap.L().Debug("decoding event", zap.Int("num_of_events", i))
			ev, err := dec.Decode()
			if err == io.EOF {
				zap.L().Info("decoded cloudevents", zap.Int("num_of_events", i))
				return nil
			}
			if err != nil {
				zap.L().Error("failed to decode cloudevent", zap.Error(err))
				return err
			}
			i += 1
			zap.L().Debug("decoded event", zap.Int("num_of_events", i))

			g.Go(func() error {
				select {
				case <-gctx.Done():
				case eventCh <- ev:
				}
				return nil
			})
		}
	}
}

func publishEvents(ctx context.Context, eventCh <-chan *event.Event, evrys evryspb.EvrysClient) func() error {
	return func() error {
		i := 0
		zap.L().Info("publishing events")
		for {
			select {
			case <-ctx.Done():
				zap.L().Info("context cancelled during publishing events", zap.Int("num_of_events", i))
				return nil
			case ev := <-eventCh:
				if ev == nil {
					zap.L().Info("published events", zap.Int("num_of_events", i))
					return nil
				}

				zap.L().Debug("publishing event", zap.Int("num_of_events", i))
				// TODO: map *event.Event to *pb.CloudEvent
				// TODO: publish event using evrys
				zap.L().Debug("published event", zap.Int("num_of_events", i))
			}
		}
	}
}
