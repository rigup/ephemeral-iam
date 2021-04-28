package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/jessesomerville/ephemeral-iam/internal/errors"
	"github.com/jessesomerville/ephemeral-iam/internal/gcpclient"
	"github.com/jessesomerville/ephemeral-iam/pkg/options"
)

var (
	kubectlCmdArgs   []string
	kubectlCmdConfig options.CmdConfig
)

func newCmdKubectl() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubectl [KUBECTL_ARGS]",
		Short: "Run a kubectl command with the permissions of the specified service account",
		Long: dedent.Dedent(`
			The "kubectl" command runs the provided kubectl command with the permissions of the specified
			service account. Output from the kubectl command is able to be piped into other commands.`),
		Example: dedent.Dedent(`
			eiam kubectl pods -o json \
			  --service-account-email example@my-project.iam.gserviceaccount.com \
			  --reason "Debugging for (JIRA-1234)"
				
			eiam kubectl pods -o json \
			  -s example@my-project.iam.gserviceaccount.com -r "example" \
			  | jq`),
		Args:               cobra.ArbitraryArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().VisitAll(options.CheckRequired)

			kubectlCmdArgs = util.ExtractUnknownArgs(cmd.Flags(), os.Args)
			if err := util.FormatReason(&kubectlCmdConfig.Reason); err != nil {
				return err
			}

			if !options.YesOption {
				util.Confirm(map[string]string{
					"Project":         kubectlCmdConfig.Project,
					"Service Account": kubectlCmdConfig.ServiceAccountEmail,
					"Reason":          kubectlCmdConfig.Reason,
					"Command":         fmt.Sprintf("kubectl %s", strings.Join(kubectlCmdArgs, " ")),
				})
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKubectlCommand()
		},
	}

	options.AddServiceAccountEmailFlag(cmd.Flags(), &kubectlCmdConfig.ServiceAccountEmail, true)
	options.AddReasonFlag(cmd.Flags(), &kubectlCmdConfig.Reason, true)
	options.AddProjectFlag(cmd.Flags(), &kubectlCmdConfig.Project)

	return cmd
}

func runKubectlCommand() error {
	hasAccess, err := gcpclient.CanImpersonate(
		kubectlCmdConfig.Project,
		kubectlCmdConfig.ServiceAccountEmail,
		kubectlCmdConfig.Reason,
	)
	if err != nil {
		return err
	} else if !hasAccess {
		util.Logger.Fatalln("You do not have access to impersonate this service account")
	}

	util.Logger.Infof("Fetching access token for %s", kubectlCmdConfig.ServiceAccountEmail)
	accessToken, err := gcpclient.GenerateTemporaryAccessToken(kubectlCmdConfig.ServiceAccountEmail, kubectlCmdConfig.Reason)
	if err != nil {
		return err
	}

	util.Logger.Infof("Running: [kubectl %s]\n\n", strings.Join(kubectlCmdArgs, " "))
	kubectlAuth := append(kubectlCmdArgs, "--token", accessToken.GetAccessToken())
	c := exec.Command(viper.GetString("binarypaths.kubectl"), kubectlAuth...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fullCmd := fmt.Sprintf("kubectl %s", strings.Join(kubectlCmdArgs, " "))
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: fmt.Sprintf("Failed to run command [%s]", fullCmd),
			Err: err,
		}
	}

	return nil
}
