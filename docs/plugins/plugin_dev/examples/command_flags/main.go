package main // All plugins must have a main package

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jessesomerville/ephemeral-iam/pkg/options"
	eiamplugin "github.com/jessesomerville/ephemeral-iam/pkg/plugins"
)

var (
	// logger will hold the logger configured by ephemeral-iam
	logger *logrus.Logger

	// Plugin is the top-level definition of the plugin.  It must be named 'Plugin'
	// and be exported by the main package
	Plugin = &eiamplugin.EphemeralIamPlugin{
		// Command defines the top-level command that will be added to eiam.
		// It is an instance of cobra.Command (https://pkg.go.dev/github.com/spf13/cobra#Command)
		Command: pluginFuncWithEiamFlags(),
		Name:    "Plugin with command flags",
		Desc:    "This is an example plugin with command flags",
		Version: "v0.0.1",
	}

	Project string
	Verbose bool
)

func init() {
	// This will instantiate logger as the same logging instance that is used
	// by eiam
	logger = eiamplugin.Logger()
}

func pluginFuncWithEiamFlags() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eiam-flags-plugin",
		Short: "Example plugin that utilizes eiam flags",
		// Plugins should use the RunE/PreRunE fields and return their errors
		// to be handled by eiam
		RunE: func(cmd *cobra.Command, args []string) error {
			if Project == "" {
				return errors.New("you must provide the `--project` flag")
			} else {
				logger.Infof("Project: %s", Project)
				if Verbose {
					logger.Info("Verbose logging is enabled")
				}
			}
			return nil
		},
	}
	// This will add the `--project` flag to the command
	options.AddProjectFlag(cmd.Flags(), &Project)
	// You can also add custom flags to the command
	cmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}
