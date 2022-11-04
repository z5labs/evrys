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
	"testing"
	"time"

	evryspb "github.com/z5labs/evrys/proto"

	"github.com/stretchr/testify/assert"
)

func TestKafkaBus(t *testing.T) {
	t.Run("will fail to publish notification", func(t *testing.T) {
		t.Run("if kafka is unavailable", func(t *testing.T) {
			bus := NewKafkaBus()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := bus.Publish(ctx, new(evryspb.Notification))
			if !assert.Error(t, err) {
				return
			}

			if !assert.IsType(t, ServiceUnavailableError{}, err) {
				return
			}

			suerr := err.(ServiceUnavailableError)
			if !assert.Equal(t, "kafka", suerr.Service) {
				return
			}
		})
	})
}
