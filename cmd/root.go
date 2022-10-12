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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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

var rootCmd = &cobra.Command{
	Use:   "evrys",
	Short: "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var lvl zapcore.Level
		lvlStr := cmd.Flags().Lookup("log-level").Value.String()
		err := lvl.UnmarshalText([]byte(lvlStr))
		if err != nil {
			panic(err)
		}

		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.OutputPaths = []string{viper.GetString("log-file")}
		l, err := cfg.Build(zap.IncreaseLevel(lvl))
		if err != nil {
			panic(err)
		}

		zap.ReplaceGlobals(l)
	},
}

func init() {
	// Persistent flags
	lvl := logLevel(zapcore.InfoLevel)
	rootCmd.PersistentFlags().Var(&lvl, "log-level", "Specify log level")
	rootCmd.PersistentFlags().String("log-file", "stderr", "Specify log file")

	viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
}
