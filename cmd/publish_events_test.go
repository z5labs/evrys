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
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestPublishEvents(t *testing.T) {
	t.Run("will return non-zero exit code", func(t *testing.T) {
		t.Run("if no args are given", func(t *testing.T) {
			err := Execute("publish", "events")
			if !assert.Error(t, err) {
				return
			}
			if !assert.Equal(t, "accepts 1 arg(s), received 0", err.Error()) {
				return
			}
		})

		t.Run("if no evrys endpoint is provided", func(t *testing.T) {
			err := Execute("publish", "events", "./testdata/events.json")
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, Error{}, err) {
				return
			}

			cmdErr := err.(Error)
			if !assert.IsType(t, UnableToDialError{}, cmdErr.Cause) {
				return
			}

			dialErr := cmdErr.Cause.(UnableToDialError)
			if !assert.Equal(t, unknownEvrysTargetErr, dialErr.Reason) {
				return
			}
		})

		t.Run("if evrys can't be dialed", func(t *testing.T) {
			err := Execute("publish", "events", "--grpc-endpoint=localhost:8080", "./testdata/events.json")
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, Error{}, err) {
				return
			}

			cmdErr := err.(Error)
			if !assert.Equal(t, grpc.ErrClientConnTimeout, cmdErr.Cause) {
				return
			}
		})
	})
}
