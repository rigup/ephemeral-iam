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
