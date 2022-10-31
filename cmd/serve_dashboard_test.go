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

func TestServeDashboard(t *testing.T) {
	t.Run("will return an error", func(t *testing.T) {
		t.Run("if any arguments are provided", func(t *testing.T) {
			err := Execute("serve", "dashboard", "foo")
			if !assert.Error(t, err) {
				return
			}
			if !assert.Equal(t, "accepts 0 arg(s), received 1", err.Error()) {
				return
			}
		})

		t.Run("if no evrys endpoint is provided", func(t *testing.T) {
			err := Execute("serve", "dashboard")
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, Error{}, err) {
				return
			}

			cmdErr := err.(Error)
			if !assert.Equal(t, "evrys serve dashboard", cmdErr.Cmd.CommandPath()) {
				return
			}
			if !assert.IsType(t, MissingEndpointError{}, cmdErr.Unwrap()) {
				return
			}
		})

		t.Run("if the evrys endpoint is unreachable", func(t *testing.T) {
			err := Execute("serve", "dashboard", "--grpc-endpoint=\"example.com:8080\"")
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, Error{}, err) {
				return
			}

			cmdErr := err.(Error)
			if !assert.Equal(t, "evrys serve dashboard", cmdErr.Cmd.CommandPath()) {
				return
			}
			if !assert.IsType(t, UnableToDialError{}, cmdErr.Unwrap()) {
				return
			}
		})
	})
}
