// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	eiamplugin "github.com/rigup/ephemeral-iam/pkg/plugins"
)

var (
	logger *logrus.Logger

	// Plugin does not need to include a RunE field
	Plugin = &eiamplugin.EphemeralIamPlugin{
		Command: &cobra.Command{
			Use:   "plugin-with-formatted-logging",
			Short: "This plugin uses the eiam logging format",
			RunE: func(cmd *cobra.Command, args []string) error {
				logger.Trace("This is a trace level log and will only be printed when the user sets 'logging.level' to 'trace'.")
				logger.Debug("This is a debug level log and will only be printed when the user sets 'logging.level' to 'trace' or 'debug'.")
				logger.Info("This is an info level log. This is the default configured log level.")
				logger.Warn("This is a warn level log.")
				logger.Error("This is an error level log. This is the highest level of log that will not halt execution.")
				logger.WithError(errors.New("something happened")).Error("You can use 'WithError' to print errors with messages.")
				logger.Fatal("This is a fatal level log. This is halt execution.")
				logger.Panic("This is a panic level log. This won't be printed because the Fatal log will stop the command.")
				return nil
			},
		},
		Name:    "Output logging example",
		Desc:    "Plugin demonstrating how to use the eiam logger",
		Version: "v0.0.1",
	}
)

func init() {
	logger = eiamplugin.Logger()
}
