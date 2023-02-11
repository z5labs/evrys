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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

type logLevel zapcore.Level

func (l logLevel) String() string {
	return (zapcore.Level)(l).String()
}

func (l *logLevel) Set(s string) error {
	return (*zapcore.Level)(l).Set(s)
}

func (l logLevel) Type() string {
	return "Level"
}

func buildEventLogCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "evrys",
		Short:             "",
		SilenceErrors:     true,
		PersistentPreRunE: withPersistentPreRun()(v),
	}

	lvl := logLevel(zapcore.InfoLevel)
	cmd.PersistentFlags().Var(&lvl, "log-level", "Specify log level")
	cmd.PersistentFlags().String("log-file", "stderr", "Specify log file")

	return cmd
}
