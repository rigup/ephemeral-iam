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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rigup/ephemeral-iam/pkg/options"
	eiamplugin "github.com/rigup/ephemeral-iam/pkg/plugins"
)

var (
	// Plugin does not need to include a RunE field.
	Plugin = &eiamplugin.EphemeralIamPlugin{
		Command: &cobra.Command{
			Use:   "plugin-with-subcommands",
			Short: "This plugin contains subcommands",
		},
		Name:    "Subcommands Example",
		Desc:    "Plugin demonstrating how to add subcommands to an eiam plugin",
		Version: "v0.0.1",
	}

	logger *logrus.Logger

	project string
)

func init() {
	logger = eiamplugin.Logger()

	Plugin.AddCommand(newCmdExampleSubcommand())
	Plugin.AddCommand(newCmdAnotherSubcommand())
}

func newCmdExampleSubcommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "example-subcommand",
		Short: "This is a subcommand of the plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Infof("We can access the current user's project even if the flag isn't provided: %s", project)
			return nil
		},
	}
	options.AddProjectFlag(cmd.Flags(), &project)
	return cmd
}

func newCmdAnotherSubcommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "another-subcommand",
		Short: "This is another subcommand of the plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Infof("The user's project isn't available to this command because the flag was not explicitly added to it: %s", project)
			return nil
		},
	}
	return cmd
}
