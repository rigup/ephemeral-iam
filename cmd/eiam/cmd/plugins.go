package cmd

import (
	"github.com/spf13/cobra"
)

func newCmdPlugins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "Manage ephemeral-iam plugins",
	}

	cmd.AddCommand(newCmdPluginsList())
	cmd.AddCommand(newCmdPluginsInstall())
	cmd.AddCommand(newCmdPluginsRemove())
	return cmd
}

func newCmdPluginsList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show the list of loaded plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			RootCommand.PrintPlugins()
			return nil
		},
	}
	return cmd
}

func newCmdPluginsInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "NOT IMPLEMENTED Install a new eiam plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func newCmdPluginsRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "NOT IMPLEMENTED Remove an installed eiam plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
