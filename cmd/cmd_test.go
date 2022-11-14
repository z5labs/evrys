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
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInitLogging(t *testing.T) {
	t.Run("will fail to initialize logger", func(t *testing.T) {
		t.Run("if a unknown logging level is set", func(t *testing.T) {
			cmd := &cobra.Command{}

			s := cmd.Flags().String("log-level", "", "")
			*s = "test"

			err := initLogging(viper.New())(cmd, nil)
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, UnknownLogLevelError{}, err) {
				return
			}

			lvlErr := err.(UnknownLogLevelError)
			if !assert.Equal(t, *s, lvlErr.Level) {
				return
			}
		})

		t.Run("if its unable to open the specified log file", func(t *testing.T) {
			cmd := &cobra.Command{}

			cmd.Flags().String("log-level", "", "")
			logFileName := cmd.Flags().String("log-file", "", "")
			*logFileName = "test.log"

			err := initLogging(viper.New())(cmd, nil)
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, UnableToInitializeLoggerError{}, err) {
				return
			}
		})
	})
}

func TestLoadConfigFile(t *testing.T) {
	t.Run("will not load config file", func(t *testing.T) {
		t.Run("if the config file does not exist", func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
			}
			cmd.Flags().String("config-file", "", "")
			cmd.Flag("config-file").Changed = true

			v := viper.New()

			err := loadConfigFile(v)(cmd, nil)
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, UnableToLoadConfigFileError{}, err) {
				return
			}

			loadErr := err.(UnableToLoadConfigFileError)
			if !assert.IsType(t, viper.ConfigFileNotFoundError{}, loadErr.Cause) {
				return
			}
		})

		t.Run("if the config file is not formatted as yaml", func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
			}
			cmd.Flags().String("config-file", "testdata/config.txt", "")
			cmd.Flag("config-file").Changed = true

			v := viper.New()

			err := loadConfigFile(v)(cmd, nil)
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, UnableToLoadConfigFileError{}, err) {
				return
			}

			loadErr := err.(UnableToLoadConfigFileError)
			if !assert.IsType(t, viper.ConfigParseError{}, loadErr.Cause) {
				return
			}
		})
	})

	t.Run("will load config file", func(t *testing.T) {
		t.Run("if the config file is formatted as yaml", func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
			}
			cmd.Flags().String("config-file", "testdata/config.yaml", "")
			cmd.Flag("config-file").Changed = true

			v := viper.New()

			err := loadConfigFile(v)(cmd, nil)
			if !assert.Nil(t, err) {
				return
			}

			s := v.GetString("hello")
			if !assert.Equal(t, "world", s) {
				return
			}
		})

		t.Run("if the config file is formatted as json", func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
			}
			cmd.Flags().String("config-file", "testdata/config.json", "")
			cmd.Flag("config-file").Changed = true

			v := viper.New()

			err := loadConfigFile(v)(cmd, nil)
			if !assert.Nil(t, err) {
				return
			}

			s := v.GetString("hello")
			if !assert.Equal(t, "world", s) {
				return
			}
		})
	})
}
