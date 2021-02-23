package cmd

import (
	"github.com/spf13/cobra"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

const version = "0.0.dev1" // TODO: Have CI set this when a release is made

func newCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the installed ephemeral-iam version",
		Run: func(cmd *cobra.Command, args []string) {
			util.Logger.Infof("ephemeral-iam v%s\n", version)
		},
	}
	return cmd
}
