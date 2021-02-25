package cmd

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/cmd/options"
)

// NewEphemeralIamCommand returns cobra.Command to run eiam command
func NewEphemeralIamCommand() *cobra.Command {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	cmds := &cobra.Command{
		Use:   "eiam",
		Short: "Utility for granting short-lived, privileged access to GCP APIs.",
		Long: dedent.Dedent(`
			╭────────────────────────────────────────────────────────────╮
			│                                                            │
			│                        Ephemeral IAM                       │
			│  ──────────────────────────────────────────────────────    │
			│  A CLI tool for temporarily escalating GCP IAM privileges  │
			│  to perform high privilege tasks.                          │
			│                                                            │
			│      https://github.com/jessesomerville/ephemeral-iam      │
			│                                                            │
			╰────────────────────────────────────────────────────────────╯
			
			
			╭────────────────────── Example usage ───────────────────────╮
			│                                                            │
			│                   Start privleged session                  │
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
			│      -r "Reason"                                           │
			│                                                            │
			╰────────────────────────────────────────────────────────────╯
		`),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmds.ResetFlags()

	cmds.AddCommand(newCmdAssumePrivileges())
	cmds.AddCommand(newCmdConfig())
	cmds.AddCommand(newCmdGcloud())
	cmds.AddCommand(newCmdKubectl())
	cmds.AddCommand(newCmdListServiceAccounts())
	cmds.AddCommand(newCmdQueryPermissions())
	options.AddPersistentFlags(cmds.PersistentFlags())

	return cmds
}
