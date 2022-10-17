/*
 * Copyright 2022 Z5Labs and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

type Error struct {
	Cmd   *cobra.Command
	Cause error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Cmd.Use, e.Cause)
}

func (e Error) Unwrap() error {
	return e.Cause
}

// CheckError
func CheckError(err error) {
	if err == nil {
		os.Exit(0)
	}
	os.Exit(1)
}

// Execute
func Execute(args ...string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if len(args) == 0 {
		args = os.Args[1:]
	}
	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(ctx)
}
