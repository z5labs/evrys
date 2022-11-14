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
	"errors"
	"net"

	"github.com/z5labs/evrys/grpc"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func withServeGrpcCmd() func(*viper.Viper) *cobra.Command {
	return func(v *viper.Viper) *cobra.Command {
		cmd := &cobra.Command{
			Use:   "grpc",
			Short: "Serve requests over gRPC",
			PersistentPreRunE: withPersistentPreRun(
				loadConfigFile(v),
			)(v),
			Run: func(cmd *cobra.Command, args []string) {
				addr := v.GetString("addr")
				ls, err := net.Listen("tcp", addr)
				if err != nil {
					zap.L().Fatal(
						"unexpected error when trying to listen on address",
						zap.String("addr", addr),
						zap.Error(err),
					)
					return
				}
				zap.L().Info("listening for grpc requests", zap.String("addr", addr))

				evrys := grpc.NewEvrysService(nil, zap.L())
				err = evrys.Serve(cmd.Context(), ls)
				if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
					zap.L().Fatal(
						"unexpected error when serving grpc traffic",
						zap.String("addr", addr),
						zap.Error(err),
					)
					return
				}
			},
		}

		return cmd
	}
}
