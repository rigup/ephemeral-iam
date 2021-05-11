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

package main // All plugins must have a main package.

import (
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/rigup/ephemeral-iam/pkg/options"
)

const (
	name    = "plugin-using-eiam-flags"
	desc    = "An example of an eiam plugin command that uses flags provided by eiam"
	version = "v0.0.1"
)

var (
	project string
	verbose bool
)

// BasicPlugin is the implementation of the ephemeral-iam plugin interface.
type BasicPlugin struct {
	// Logger is the logger for the plugin to use to send log entries to eiam
	// to be formatted and output to the user.
	Logger hclog.Logger
}

// GetInfo is the function that eiam invokes to get metadata about the plugin.
func (p *BasicPlugin) GetInfo() (n, d, v string, err error) {
	return name, desc, version, nil
}

// Run is the function that eiam uses to invoke the plugin command.
func (p *BasicPlugin) Run() error {
	rootCmd := newRootCmd(p)
	return rootCmd.Execute()
}

func newRootCmd(p *BasicPlugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: desc,
		// Plugins should use the RunE/PreRunE fields and return their errors
		// to be handled by eiam.
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Logger.Info("The project field defaults to the value in the user's gcloud config", "project", project)
			if verbose {
				p.Logger.Info("Verbose logging is enabled")
			}
			return nil
		},
	}
	// This will add the `--project` flag to the command.
	options.AddProjectFlag(cmd.Flags(), &project)
	// You can also add custom flags to the command.
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}
