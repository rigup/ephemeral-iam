package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/rigup/ephemeral-iam/pkg/options"
	eiamplugin "github.com/rigup/ephemeral-iam/pkg/plugins"
)

var (
	// Plugin does not need to include a RunE field
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

	Project string
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
			logger.Infof("We can access the current user's project even if the flag isn't provided: %s", Project)
			return nil
		},
	}
	options.AddProjectFlag(cmd.Flags(), &Project)
	return cmd
}

func newCmdAnotherSubcommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "another-subcommand",
		Short: "This is another subcommand of the plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Infof("The user's project isn't available to this command because the flag was not explicitly added to it: %s", Project)
			return nil
		},
	}
	return cmd
}
