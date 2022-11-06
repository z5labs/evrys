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

package notification

import (
	"context"
	"net"
	"testing"
	"time"

	evryspb "github.com/z5labs/evrys/proto"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestKafkaBus(t *testing.T) {
	t.Run("will fail to initialize", func(t *testing.T) {
		t.Run("if no logger is provided", func(t *testing.T) {
			newKafkaBus := func(cfg KafkaConfig) (err error) {
				defer func() {
					rerr := recover()
					if !assert.NotNil(t, rerr) {
						return
					}

					e, ok := rerr.(error)
					if !assert.True(t, ok) {
						return
					}
					err = e
				}()

				NewKafkaBus(cfg)
				return
			}

			err := newKafkaBus(KafkaConfig{
				Addresses: []string{""},
			})
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, ConfigurationError{}, err) {
				return
			}

			cfgErr := err.(ConfigurationError)
			if !assert.Equal(t, "kafka", cfgErr.Bus) {
				return
			}
			if !assert.Equal(t, ErrMissingLogger, cfgErr.Cause) {
				return
			}
		})

		t.Run("if no addresses are provided", func(t *testing.T) {
			newKafkaBus := func(cfg KafkaConfig) (err error) {
				defer func() {
					rerr := recover()
					if !assert.NotNil(t, rerr) {
						return
					}

					e, ok := rerr.(error)
					if !assert.True(t, ok) {
						return
					}
					err = e
				}()

				NewKafkaBus(cfg)
				return
			}

			err := newKafkaBus(KafkaConfig{
				Logger: zap.L(),
			})
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, ConfigurationError{}, err) {
				return
			}

			cfgErr := err.(ConfigurationError)
			if !assert.Equal(t, "kafka", cfgErr.Bus) {
				return
			}
			if !assert.Equal(t, ErrMissingKafkaAddrs, cfgErr.Cause) {
				return
			}
		})
	})

	t.Run("will fail to publish notification", func(t *testing.T) {
		t.Run("if given an invalid notification", func(t *testing.T) {
			bus := NewKafkaBus(KafkaConfig{
				Addresses: []string{"localhost:10000"},
				Logger:    zap.L(),
			})

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			n := &evryspb.Notification{}
			err := bus.Publish(ctx, n)
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, ValidationError{}, err) {
				return
			}
		})

		t.Run("if kafka is unavailable", func(t *testing.T) {
			bus := NewKafkaBus(KafkaConfig{
				Addresses: []string{"localhost:10000"},
				Logger:    zap.L(),
			})

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			n := &evryspb.Notification{
				EventSource: "source",
				EventId:     "id",
				EventType:   "type",
			}
			err := bus.Publish(ctx, n)
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, Error{}, err) {
				return
			}

			busErr := err.(Error)
			if !assert.Equal(t, "kafka", busErr.Bus) {
				return
			}
			if !assert.Error(t, busErr.Cause) {
				return
			}
			if !assert.IsType(t, &net.OpError{}, busErr.Cause) {
				t.Log(busErr.Cause)
				return
			}
		})
	})
}
