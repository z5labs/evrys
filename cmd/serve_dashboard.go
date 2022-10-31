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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func withServeDashboardCmd() func(*viper.Viper) *cobra.Command {
	return func(v *viper.Viper) *cobra.Command {
		cmd := &cobra.Command{
			Use:   "dashboard",
			Short: "Serve a web based dashboard for easily interacting with evrys",
			Args:  cobra.ExactArgs(0),
			RunE:  serveDashboard(v),
		}

		cmd.Flags().String("grpc-endpoint", "", "gRPC endpoint of evrys service")

		return cmd
	}
}

func serveDashboard(v *viper.Viper) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		e := getEndpoint(v)
		_, cleanupConn, err := dialEvrys(cmd.Context(), e)
		if err != nil {
			return Error{
				Cmd:   cmd,
				Cause: err,
			}
		}
		defer cleanupConn()

		return nil
	}
}
