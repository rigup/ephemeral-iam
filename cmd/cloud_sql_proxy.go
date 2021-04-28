package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
	"github.com/rigup/ephemeral-iam/internal/gcpclient"
	"github.com/rigup/ephemeral-iam/pkg/options"
)

var (
	cloudSqlProxyCmdArgs   []string
	cloudSqlProxyCmdConfig options.CmdConfig
)

func newCmdCloudSqlProxy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud_sql_proxy [GCLOUD_ARGS]",
		Short: "Run cloud_sql_proxy with the permissions of the specified service account",
		Long: dedent.Dedent(`
			The "cloud_sql_proxy" command runs the provided cloud_sql_proxy command with the permissions of the specified
			service account.`),
		Example: dedent.Dedent(`
			eiam cloud_sql_proxy -instances "example_db_instance" \
			--service-account-email example@my-project.iam.gserviceaccount.com \
			--reason "Debugging for (JIRA-1234)"`),
		Args:               cobra.ArbitraryArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if viper.GetString("binarypaths.cloudSqlProxy") == "" {
				err := errors.New(`"cloud_sql_proxy": executable file not found in $PATH`)
				return errorsutil.EiamError{
					Log: util.Logger.WithError(err),
					Msg: "Failed to run cloud_sql_proxy command",
					Err: err,
				}
			}

			cmd.Flags().VisitAll(options.CheckRequired)

			cloudSqlProxyCmdArgs = util.ExtractUnknownArgs(cmd.Flags(), os.Args)
			if err := util.FormatReason(&cloudSqlProxyCmdConfig.Reason); err != nil {
				return err
			}

			if !options.YesOption {
				util.Confirm(map[string]string{
					"Project":         cloudSqlProxyCmdConfig.Project,
					"Service Account": cloudSqlProxyCmdConfig.ServiceAccountEmail,
					"Reason":          cloudSqlProxyCmdConfig.Reason,
					"Command":         fmt.Sprintf("cloud_sql_proxy %s", strings.Join(cloudSqlProxyCmdArgs, " ")),
				})
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCloudSqlProxyCommand()
		},
	}

	options.AddServiceAccountEmailFlag(cmd.Flags(), &cloudSqlProxyCmdConfig.ServiceAccountEmail, true)
	options.AddReasonFlag(cmd.Flags(), &cloudSqlProxyCmdConfig.Reason, true)
	options.AddProjectFlag(cmd.Flags(), &cloudSqlProxyCmdConfig.Project)

	return cmd
}

func runCloudSqlProxyCommand() error {
	hasAccess, err := gcpclient.CanImpersonate(
		cloudSqlProxyCmdConfig.Project,
		cloudSqlProxyCmdConfig.ServiceAccountEmail,
		cloudSqlProxyCmdConfig.Reason,
	)
	if err != nil {
		return err
	} else if !hasAccess {
		util.Logger.Fatalln("You do not have access to impersonate this service account")
	}

	util.Logger.Infof("Fetching access token for %s", cloudSqlProxyCmdConfig.ServiceAccountEmail)
	accessToken, err := gcpclient.GenerateTemporaryAccessToken(cloudSqlProxyCmdConfig.ServiceAccountEmail, cloudSqlProxyCmdConfig.Reason)
	if err != nil {
		return err
	}

	util.Logger.Infof("Running: [cloud_sql_proxy %s]\n\n", strings.Join(cloudSqlProxyCmdArgs, " "))
	cloudSqlProxyAuth := append(cloudSqlProxyCmdArgs, "-token", accessToken.GetAccessToken())
	c := exec.Command(viper.GetString("binarypaths.cloudSqlProxy"), cloudSqlProxyAuth...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fullCmd := fmt.Sprintf("cloud_sql_proxy %s", strings.Join(cloudSqlProxyCmdArgs, " "))
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: fmt.Sprintf("Failed to run command [%s]", fullCmd),
			Err: err,
		}
	}
	return nil
}
