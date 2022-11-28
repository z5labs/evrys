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

package testcontainer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKafkaCluster(t *testing.T) {
	t.Run("cannot be constructed", func(t *testing.T) {
		t.Run("if its misconfigured", func(t *testing.T) {
			testCases := []struct {
				Name string
				Opts []KafkaOption
				Err  error
			}{
				{
					Name: "not full semantic docker tag version",
					Opts: []KafkaOption{
						WithKafkaTag("1.2"),
					},
					Err: ErrInvalidKafkaTag,
				},
				{
					Name: "non-numeric semantic version element in docker tag",
					Opts: []KafkaOption{
						WithKafkaTag("1.1.a"),
					},
					Err: ErrInvalidKafkaTag,
				},
				{
					Name: "empty cluster network name",
					Opts: []KafkaOption{
						WithClusterNetworkName(""),
					},
					Err: ErrInvalidKafkaNetworkName,
				},
				{
					Name: "all whitespace cluster network name",
					Opts: []KafkaOption{
						WithClusterNetworkName("     "),
					},
					Err: ErrInvalidKafkaNetworkName,
				},
				{
					Name: "empty zookeeper port",
					Opts: []KafkaOption{
						WithZooKeeperPort(""),
					},
					Err: ErrInvalidZookeeperPort,
				},
				{
					Name: "non-integer zookeeper port",
					Opts: []KafkaOption{
						WithZooKeeperPort("a"),
					},
					Err: ErrInvalidZookeeperPort,
				},
				{
					Name: "empty kafka broker port",
					Opts: []KafkaOption{
						WithKafkaBrokerPort(""),
					},
					Err: ErrInvalidKafkaBrokerPort,
				},
				{
					Name: "non-integer kafka broker port",
					Opts: []KafkaOption{
						WithKafkaBrokerPort("a"),
					},
					Err: ErrInvalidKafkaBrokerPort,
				},
				{
					Name: "empty kafka client port",
					Opts: []KafkaOption{
						WithKafkaClientPort(""),
					},
					Err: ErrInvalidKafkaClientPort,
				},
				{
					Name: "non-integer kafka client port",
					Opts: []KafkaOption{
						WithKafkaClientPort("a"),
					},
					Err: ErrInvalidKafkaClientPort,
				},
			}

			for _, testCase := range testCases {
				t.Run(testCase.Name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					_, err := NewKafkaCluster(ctx, testCase.Opts...)
					if !assert.Error(t, err) {
						return
					}
					if !assert.IsType(t, ValidationError{}, err) {
						return
					}

					verr := err.(ValidationError)
					if !assert.Equal(t, testCase.Err, verr.Cause) {
						return
					}
				})
			}
		})

		// TODO: figure out how to get the rest of the constructor code to fail
	})

	t.Run("will fail to start", func(t *testing.T) {
		t.Run("if the zookeeper container fails to start", func(t *testing.T) {
			t.Fail() // TODO
		})

		t.Run("if the kafka container fails to start", func(t *testing.T) {
			t.Fail() // TODO
		})

		t.Run("if it can not create a temp file", func(t *testing.T) {
			t.Fail() // TODO
		})

		t.Run("if it can not get the kafka host", func(t *testing.T) {
			t.Fail() // TODO
		})

		t.Run("if it can not copy the temp file to the kafka container", func(t *testing.T) {
			t.Fail() // TODO
		})
	})

	t.Run("will successfully start", func(t *testing.T) {
		t.Run("if using the default config", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			kafka, err := NewKafkaCluster(ctx)
			if !assert.Nil(t, err) {
				return
			}

			err = kafka.Start(ctx)
			if !assert.Nil(t, err) {
				return
			}

			err = kafka.Terminate(ctx)
			if !assert.Nil(t, err) {
				return
			}
		})
	})
}
