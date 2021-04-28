package cmd

import (
	"github.com/spf13/cobra"

	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
)

func newCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the installed ephemeral-iam version",
		Run: func(cmd *cobra.Command, args []string) {
			util.Logger.Infof("ephemeral-iam %s\n", appconfig.Version)
		},
	}
	return cmd
}
