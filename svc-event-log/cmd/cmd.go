// Copyright 2023 Z5Labs and Contributors
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
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	return ExecuteContext(context.Background(), args...)
}

// ExecuteContext
func ExecuteContext(pctx context.Context, args ...string) error {
	ctx, cancel := signal.NotifyContext(pctx, os.Interrupt, os.Kill)
	defer cancel()

	if len(args) == 0 {
		args = os.Args[1:]
	}

	cmd := buildCli(
		withServeCommand(
			withServeGrpcCmd(),
		),
	)
	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}

func buildCli(subcommandBuilders ...func(v *viper.Viper) *cobra.Command) *cobra.Command {
	v := viper.New()
	cmd := buildEventLogCmd(v)

	for _, b := range subcommandBuilders {
		cmd.AddCommand(b(v))
	}
	return cmd
}

type preRunFunc func(*cobra.Command, []string) error

func withPersistentPreRun(fs ...preRunFunc) func(*viper.Viper) preRunFunc {
	return func(v *viper.Viper) preRunFunc {
		preRuns := []preRunFunc{
			bindFlags(v),
			initLogging(v),
		}
		preRuns = append(preRuns, fs...)

		return func(cmd *cobra.Command, args []string) error {
			for _, f := range preRuns {
				err := f(cmd, args)
				if err != nil {
					return Error{
						Cmd:   cmd,
						Cause: err,
					}
				}
			}
			return nil
		}
	}
}

func bindFlags(v *viper.Viper) preRunFunc {
	return func(cmd *cobra.Command, args []string) error {
		v.BindPFlags(cmd.Flags())
		v.BindPFlags(cmd.PersistentFlags())
		return nil
	}
}

// UnknownLogLevelError
type UnknownLogLevelError struct {
	Level string
}

func (e UnknownLogLevelError) Error() string {
	return fmt.Sprintf("unknown logging level: %s", e.Level)
}

// UnableToInitializeLoggerError
type UnableToInitializeLoggerError struct {
	Cause error
}

func (e UnableToInitializeLoggerError) Error() string {
	return fmt.Sprintf("failed to initialize logger: %s", e.Cause)
}

func (e UnableToInitializeLoggerError) Unwrap() error {
	return e.Cause
}

func initLogging(v *viper.Viper) preRunFunc {
	return func(cmd *cobra.Command, args []string) error {
		var lvl zapcore.Level
		lvlStr := cmd.Flags().Lookup("log-level").Value.String()
		err := lvl.UnmarshalText([]byte(lvlStr))
		if err != nil {
			return UnknownLogLevelError{
				Level: lvlStr,
			}
		}

		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.OutputPaths = []string{v.GetString("log-file")}
		l, err := cfg.Build(zap.IncreaseLevel(lvl))
		if err != nil {
			return UnableToInitializeLoggerError{
				Cause: err,
			}
		}

		zap.ReplaceGlobals(l)
		return nil
	}
}

// UnableToLoadConfigFileError signifies any issue that may come up when trying to load a config file.
type UnableToLoadConfigFileError struct {
	Cause error
}

func (e UnableToLoadConfigFileError) Error() string {
	return fmt.Sprintf("failed to load config file: %s", e.Cause)
}

func (e UnableToLoadConfigFileError) Unwrap() error {
	return e.Cause
}

func loadConfigFile(v *viper.Viper) preRunFunc {
	return func(cmd *cobra.Command, args []string) error {
		flag := cmd.Flag("config-file")
		if flag == nil {
			return nil
		}
		if !flag.Changed {
			return nil
		}

		v.SetConfigFile(flag.Value.String())
		v.SetConfigType("yaml")
		err := v.ReadInConfig()
		if err != nil {
			return UnableToLoadConfigFileError{
				Cause: err,
			}
		}
		return nil
	}
}
