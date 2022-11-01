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

func TestLoadConfigFile(t *testing.T) {
	t.Run("will not load config file", func(t *testing.T) {
		t.Run("if the config file does not exist", func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
			}
			cmd.Flags().String("config-file", "", "")
			cmd.Flag("config-file").Changed = true

			v := viper.New()

			f := func() (err error) {
				defer func() {
					r := recover()
					if !assert.NotNil(t, r) {
						return
					}

					e, ok := r.(error)
					if !assert.True(t, ok) {
						return
					}
					err = e
				}()

				loadConfigFile(v)(cmd, nil)
				return
			}

			err := f()
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, viper.ConfigFileNotFoundError{}, err) {
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

			f := func() (err error) {
				defer func() {
					r := recover()
					if !assert.NotNil(t, r) {
						return
					}

					e, ok := r.(error)
					if !assert.True(t, ok) {
						return
					}
					err = e
				}()

				loadConfigFile(v)(cmd, nil)
				return
			}

			err := f()
			if !assert.Error(t, err) {
				return
			}
			if !assert.IsType(t, viper.ConfigParseError{}, err) {
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

			f := func() (err error) {
				defer func() {
					r := recover()
					if assert.Nil(t, r) {
						return
					}

					e, ok := r.(error)
					if !ok {
						return
					}
					err = e
				}()

				loadConfigFile(v)(cmd, nil)
				return
			}

			err := f()
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

			f := func() (err error) {
				defer func() {
					r := recover()
					if assert.Nil(t, r) {
						return
					}

					e, ok := r.(error)
					if !ok {
						return
					}
					err = e
				}()

				loadConfigFile(v)(cmd, nil)
				return
			}

			err := f()
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
