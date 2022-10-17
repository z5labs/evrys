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
)

func TestPublishEvents(t *testing.T) {
	t.Run("will return non-zero exit code", func(t *testing.T) {
		t.Run("if no args are given", func(t *testing.T) {
			err := Execute("publish", "events")
			if !assert.Error(t, err) {
				return
			}
			if !assert.Equal(t, "requires at least 1 arg(s), only received 0", err.Error()) {
				return
			}
		})

		t.Run("if evrys endpoint in unreachable", func(t *testing.T) {
			err := Execute("publish", "events", "\"{}\"", "--grpc-endpoint=\"example.com:8080\"")
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, Error{}, err) {
				return
			}

			cmdErr := err.(Error)
			if !assert.Equal(t, publishEventsCmd.Use, cmdErr.Cmd.Use) {
				return
			}
			if !assert.IsType(t, UnableToDialError{}, cmdErr.Unwrap()) {
				return
			}
		})
	})
}
