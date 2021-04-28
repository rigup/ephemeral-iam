package cmd

import (
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	eiam "github.com/jessesomerville/ephemeral-iam/internal"
	"github.com/jessesomerville/ephemeral-iam/pkg/options"
)

var RootCommand *eiam.RootCommand

// NewEphemeralIamCommand returns cobra.Command to run eiam command
func NewEphemeralIamCommand() (*eiam.RootCommand, error) {
	cmds := &eiam.RootCommand{Command: cobra.Command{
		Use:   "eiam",
		Short: "Utility for granting short-lived, privileged access to GCP APIs.",
		Long: dedent.Dedent(`
			╭────────────────────────────────────────────────────────────╮
			│                                                            │
			│                        ephemeral-iam                       │
			│  ──────────────────────────────────────────────────────    │
			│  A CLI tool for temporarily escalating GCP IAM privileges  │
			│  to perform high privilege tasks.                          │
			│                                                            │
			│      https://github.com/jessesomerville/ephemeral-iam      │
			│                                                            │
			╰────────────────────────────────────────────────────────────╯
			
			
			╭────────────────────── Example usage ───────────────────────╮
			│                                                            │
			│                   Start privileged session                 │
			│  ──────────────────────────────────────────────────────    │
			│  $ eiam assumePrivileges \                                 │
			│      -s example-svc@my-project.iam.gserviceaccount.com \   │
			│      --reason "Emergency security patch (JIRA-1234)"       │
			│                                                            │
			│                                                            │
			│                                                            │
			│                     Run gcloud command                     │
			│  ──────────────────────────────────────────────────────    │
			│  $ eiam gcloud compute instances list --format=json \      │
			│      -s example@my-project.iam.gserviceaccount.com \       │
			│      -R "Reason"                                           │
			│                                                            │
			╰────────────────────────────────────────────────────────────╯
		
			Please report any bugs or feature requests by opening a new
			issue at https://github.com/jessesomerville/ephemeral-iam/issues
		`),
		SilenceErrors: true,
		SilenceUsage:  true,
	}}

	cmds.ResetFlags()

	cmds.AddCommand(newCmdAssumePrivileges())
	cmds.AddCommand(newCmdCloudSqlProxy())
	cmds.AddCommand(newCmdConfig())
	cmds.AddCommand(newCmdGcloud())
	cmds.AddCommand(newCmdKubectl())
	cmds.AddCommand(newCmdListServiceAccounts())
	cmds.AddCommand(newCmdPlugins())
	cmds.AddCommand(newCmdQueryPermissions())
	cmds.AddCommand(newCmdVersion())
	if err := cmds.LoadPlugins(); err != nil {
		return nil, err
	}
	options.AddPersistentFlags(cmds.PersistentFlags())

	RootCommand = cmds

	return cmds, nil
}
