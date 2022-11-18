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

// Package testcontainer provides some helpful abstractions over the testcontainers package.
package testcontainer

import "fmt"

// ValidationError
type ValidationError struct {
	Cause error
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("testcontainer: must provide a valid notification: %s", e.Cause)
}

func (e ValidationError) Unwrap() error {
	return e.Cause
}

// FailedToCreateNetwork
type FailedToCreateNetwork struct {
	Name  string
	Cause error
}

func (e FailedToCreateNetwork) Error() string {
	return fmt.Sprintf("%s: failed to create network: %s", e.Name, e.Cause)
}

func (e FailedToCreateNetwork) Unwrap() error {
	return e.Cause
}

// FailedToCreateContainer
type FailedToCreateContainer struct {
	Image string
	Cause error
}

func (e FailedToCreateContainer) Error() string {
	return fmt.Sprintf("%s: failed to create container: %s", e.Image, e.Cause)
}

func (e FailedToCreateContainer) Unwrap() error {
	return e.Cause
}

// FailedToStartContainer
type FailedToStartContainer struct {
	Image string
	Cause error
}

func (e FailedToStartContainer) Error() string {
	return fmt.Sprintf("%s: failed to start container: %s", e.Image, e.Cause)
}

func (e FailedToStartContainer) Unwrap() error {
	return e.Cause
}
