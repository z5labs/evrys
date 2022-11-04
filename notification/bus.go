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
	"fmt"

	evryspb "github.com/z5labs/evrys/proto"
)

// Bus
type Bus interface {
	// Publish
	Publish(context.Context, *evryspb.Notification) error
}

// ServiceUnavailableError
type ServiceUnavailableError struct {
	Service string
}

// Error
func (e ServiceUnavailableError) Error() string {
	return fmt.Sprintf("%s: service unavailable", e.Service)
}
