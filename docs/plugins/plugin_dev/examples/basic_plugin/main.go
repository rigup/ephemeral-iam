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

package main // All plugins must have a main package

import (
	"errors"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	eiamplugin "github.com/rigup/ephemeral-iam/pkg/plugins"
)

var (
	// logger will hold the logger configured by ephemeral-iam
	logger *logrus.Logger

	// Plugin is the top-level definition of the plugin.  It must be named 'Plugin'
	// and be exported by the main package
	Plugin = &eiamplugin.EphemeralIamPlugin{
		// Command defines the top-level command that will be added to eiam.
		// It is an instance of cobra.Command (https://pkg.go.dev/github.com/spf13/cobra#Command)
		Command: &cobra.Command{
			Use:   "basic-plugin",
			Short: "Basic plugin help message",
			// Plugins should use the RunE/PreRunE fields and return their errors
			// to be handled by eiam
			RunE: func(cmd *cobra.Command, args []string) error {
				logger.Info("This is printed in the same format as other `eiam` INFO logs")
				logger.Error("This is an error message")
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(2) == 1 {
					return errors.New("this is an example error returned to eiam")
				}
				return nil
			},
		},
		Name:    "Basic Plugin",
		Desc:    "This is a basic single command plugin",
		Version: "v0.0.1",
	}
)

func init() {
	// This will instantiate logger as the same logging instance that is used
	// by eiam
	logger = eiamplugin.Logger()
}
