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
	"testing"

	"github.com/stretchr/testify/assert"
	evryspb "github.com/z5labs/evrys/proto"
)

func TestValidateNotification(t *testing.T) {
	t.Run("will report a notification as invalid", func(t *testing.T) {
		testCases := []struct {
			Condition string
			N         *evryspb.Notification
			Err       error
		}{
			// Event Type cases
			{
				Condition: "if the event type is not set",
				N: &evryspb.Notification{
					EventSource: "source",
					EventId:     "id",
				},
				Err: ErrInvalidEventType,
			},
			{
				Condition: "if the event type is an empty string",
				N: &evryspb.Notification{
					EventSource: "source",
					EventId:     "id",
					EventType:   "",
				},
				Err: ErrInvalidEventType,
			},
			{
				Condition: "if the event type is a blank string",
				N: &evryspb.Notification{
					EventSource: "source",
					EventId:     "id",
					EventType:   "      ",
				},
				Err: ErrInvalidEventType,
			},
			// Event Id cases
			{
				Condition: "if the event id is not set",
				N: &evryspb.Notification{
					EventSource: "source",
					EventType:   "type",
				},
				Err: ErrInvalidEventId,
			},
			{
				Condition: "if the event id is an empty string",
				N: &evryspb.Notification{
					EventSource: "source",
					EventId:     "",
					EventType:   "type",
				},
				Err: ErrInvalidEventId,
			},
			{
				Condition: "if the event id is a blank string",
				N: &evryspb.Notification{
					EventSource: "source",
					EventId:     "      ",
					EventType:   "type",
				},
				Err: ErrInvalidEventId,
			},
			// Event Source cases
			{
				Condition: "if the event source is not set",
				N: &evryspb.Notification{
					EventId:   "id",
					EventType: "type",
				},
				Err: ErrInvalidEventSource,
			},
			{
				Condition: "if the event source is an empty string",
				N: &evryspb.Notification{
					EventSource: "",
					EventId:     "id",
					EventType:   "type",
				},
				Err: ErrInvalidEventSource,
			},
			{
				Condition: "if the event source is a blank string",
				N: &evryspb.Notification{
					EventSource: "       ",
					EventId:     "id",
					EventType:   "type",
				},
				Err: ErrInvalidEventSource,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Condition, func(t *testing.T) {
				err := validateNotification(testCase.N)
				if !assert.Error(t, err) {
					return
				}
				if !assert.Equal(t, testCase.Err, err) {
					return
				}
			})
		}
	})

	t.Run("will report a notification as valid", func(t *testing.T) {
		t.Run("if the event id, source, and type are set to non blank or empty strings", func(t *testing.T) {
			n := &evryspb.Notification{
				EventSource: "source",
				EventId:     "id",
				EventType:   "type",
			}

			err := validateNotification(n)
			if !assert.Nil(t, err) {
				return
			}
		})
	})
}
