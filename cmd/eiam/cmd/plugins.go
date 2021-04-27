package cmd

import (
	"os"

	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
	eiamplugin "github.com/jessesomerville/ephemeral-iam/pkg/plugins"
	"github.com/lithammer/dedent"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func newCmdPlugins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "Manage ephemeral-iam plugins",
		Long: dedent.Dedent(`
			Plugins for ephemeral-iam allow you to extend eiam's functionality in the form of commands.
			Plugins are '.so' files (Golang dynamic libraries) and stored in the 'plugins' directory
			of your eiam configuration folder.
			
			- Installing a plugin -
			To install a plugin, take the '.so' file and place it in the 'plugins' directory of your
			eiam configuration folder.  eiam will automatically load any valid plugins in that
			directory and the commands added by those plugins will be listed when you run:
			'eiam --help'.
		`),
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
			if len(RootCommand.Plugins) == 0 {
				util.Logger.Warn("No plugins are currently installed")
				return nil
			}
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
			util.Logger.Error("This command is not yet implemented")
			return nil
		},
	}
	return cmd
}

func newCmdPluginsRemove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove an installed eiam plugin",
		Long: dedent.Dedent(`
			The "plugins remove" command removes a currently installed plugin.
			
			You will be prompted to select the plugin to uninstall from the list of plugins loaded
			by eiam. If no plugins are currently installed, a warning is shown.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(RootCommand.Plugins) == 0 {
				util.Logger.Warn("No plugins are currently installed")
				return nil
			}

			plugin, err := selectPlugin()
			if err != nil {
				return err
			}

			if err := os.Remove(plugin.Path); err != nil {
				return err
			}
			util.Logger.Infof("Successfully removed %s", plugin.Name)
			return nil
		},
	}
	return cmd
}

func selectPlugin() (*eiamplugin.EphemeralIamPlugin, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   " ►  {{ .Name | red }}",
		Inactive: "  {{ .Name | red }}",
		Selected: " ►  {{ .Name | red | cyan }}",
		Details: `
--------- Plugin ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Desc }}`,
	}

	prompt := promptui.Select{
		Label:     "Plugin to remove",
		Items:     RootCommand.Plugins,
		Templates: templates,
	}

	if i, _, err := prompt.Run(); err != nil {
		return nil, err
	} else {
		return RootCommand.Plugins[i], nil
	}
}
