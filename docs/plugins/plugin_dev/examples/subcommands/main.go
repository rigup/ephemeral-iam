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
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"

	"github.com/rigup/ephemeral-iam/pkg/options"
)

const (
	name    = "basic-plugin"
	desc    = "An example of a basic eiam plugin command"
	version = "v0.0.1"
)

var project string

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
	}
	cmd.AddCommand(newCmdExampleSubcommand(p))
	cmd.AddCommand(newCmdAnotherSubcommand(p))
	return cmd
}

func newCmdExampleSubcommand(p *BasicPlugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "example-subcommand",
		Short: "This is a subcommand of the plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Logger.Info("We can access the current user's project even if the flag isn't provided", "project", project)
			return nil
		},
	}
	options.AddProjectFlag(cmd.Flags(), &project)
	return cmd
}

func newCmdAnotherSubcommand(p *BasicPlugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "another-subcommand",
		Short: "This is another subcommand of the plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Logger.Info("The user's project is empty because the flag was not explicitly added to it", "project", project)
			return nil
		},
	}
	return cmd
}
