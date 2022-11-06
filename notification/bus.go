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

// Package notification provides a common abstraction over various notification bus services.
package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"

	evryspb "github.com/z5labs/evrys/proto"
)

var (
	ErrMissingLogger = errors.New("logger is required")
)

// Bus
type Bus interface {
	// Publish
	Publish(context.Context, *evryspb.Notification) error
}

// Error represents a generic error a Bus implementation encountered
type Error struct {
	Bus   string
	Cause error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: encountered unexpected error: %s", e.Bus, e.Cause)
}

func (e Error) Unwrap() error {
	return e.Cause
}

// MarshalError
type MarshalError struct {
	Protocol string
	Cause    error
}

func (e MarshalError) Error() string {
	return fmt.Sprintf("%s: failed to marshal notification: %s", e.Protocol, e.Cause)
}

func (e MarshalError) Unwrap() error {
	return e.Cause
}

// ConfigurationError
type ConfigurationError struct {
	Bus   string
	Cause error
}

func (e ConfigurationError) Error() string {
	return fmt.Sprintf("%s: invalid configuration: %s", e.Bus, e.Cause)
}

func (e ConfigurationError) Unwrap() error {
	return e.Cause
}

// ValidationError
type ValidationError struct {
	Cause error
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("notification: must provide a valid notification: %s", e.Cause)
}

func (e ValidationError) Unwrap() error {
	return e.Cause
}

var (
	ErrInvalidEventType   = errors.New("event type must be non-empty and non-blank")
	ErrInvalidEventId     = errors.New("event id must be non-empty and non-blank")
	ErrInvalidEventSource = errors.New("event source must be non-empty and non-blank")
)

type validator func(*evryspb.Notification) error

func validateNotification(n *evryspb.Notification) error {
	validators := []validator{
		validateEventType,
		validateEventId,
		validateEventSource,
	}
	for _, validator := range validators {
		err := validator(n)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateEventType(n *evryspb.Notification) error {
	eventType := strings.TrimSpace(n.EventType)
	if len(eventType) == 0 {
		return ErrInvalidEventType
	}
	return nil
}

func validateEventId(n *evryspb.Notification) error {
	eventId := strings.TrimSpace(n.EventId)
	if len(eventId) == 0 {
		return ErrInvalidEventId
	}
	return nil
}

func validateEventSource(n *evryspb.Notification) error {
	eventSource := strings.TrimSpace(n.EventSource)
	if len(eventSource) == 0 {
		return ErrInvalidEventSource
	}
	return nil
}
