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
)

func withServeCommand(subcommandBuilders ...func(*viper.Viper) *cobra.Command) func(*viper.Viper) *cobra.Command {
	return func(v *viper.Viper) *cobra.Command {
		cmd := &cobra.Command{
			Use:   "serve",
			Short: "Serve requests",
			PersistentPreRunE: withPersistentPreRun(
				loadConfigFile(v),
			)(v),
		}

		// Flags
		cmd.PersistentFlags().String("addr", "0.0.0.0:8080", "Address to listen for connections.")
		cmd.PersistentFlags().String("config-file", "", "Specify config file")

		for _, b := range subcommandBuilders {
			cmd.AddCommand(b(v))
		}

		return cmd
	}
}
