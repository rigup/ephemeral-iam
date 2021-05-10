// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eiam

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
	cloudSQLProxyCmdArgs []string
	cspCmdConfig         options.CmdConfig

	cloudSQLProxyPath string
)

func NewCmdCloudSQLProxy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud_sql_proxy [GCLOUD_ARGS]",
		Short: "Run cloud_sql_proxy with the permissions of the specified service account",
		Long: dedent.Dedent(`
			The "cloud_sql_proxy" command runs the provided cloud_sql_proxy command with the permissions of the specified
			service account.`),
		Example: dedent.Dedent(`
			eiam cloud_sql_proxy -instances my-project:us-central1:example-instance=tcp:3306 \
			--service-account-email example@my-project.iam.gserviceaccount.com \
			--reason "Debugging for (JIRA-1234)"`),
		Args:               cobra.ArbitraryArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cloudSQLProxyPath = viper.GetString("binarypaths.cloudsqlproxy"); cloudSQLProxyPath == "" {
				err := errors.New(`"cloud_sql_proxy": executable file not found in $PATH`)
				return errorsutil.New("Failed to run cloud_sql_proxy command", err)
			}

			cmd.Flags().VisitAll(options.CheckRequired)

			cloudSQLProxyCmdArgs = util.ExtractUnknownArgs(cmd.Flags(), os.Args)
			if err := util.FormatReason(&cspCmdConfig.Reason); err != nil {
				return err
			}

			if !options.YesOption {
				util.Confirm(map[string]string{
					"Project":         cspCmdConfig.Project,
					"Service Account": cspCmdConfig.ServiceAccountEmail,
					"Reason":          cspCmdConfig.Reason,
					"Command":         fmt.Sprintf("cloud_sql_proxy %s", strings.Join(cloudSQLProxyCmdArgs, " ")),
				})
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCloudSQLProxyCommand()
		},
	}

	options.AddServiceAccountEmailFlag(cmd.Flags(), &cspCmdConfig.ServiceAccountEmail, true)
	options.AddReasonFlag(cmd.Flags(), &cspCmdConfig.Reason, true)
	options.AddProjectFlag(cmd.Flags(), &cspCmdConfig.Project)

	return cmd
}

func runCloudSQLProxyCommand() error {
	hasAccess, err := gcpclient.CanImpersonate(cspCmdConfig.Project, cspCmdConfig.ServiceAccountEmail)
	if err != nil {
		return err
	} else if !hasAccess {
		return errorsutil.EiamError{
			Log: util.Logger.WithField("cmd", "cloud_sql_proxy"),
			Msg: "You do not have access to impersonate this service account",
			Err: errors.New("failed to impersonate service account"),
		}
	}

	util.Logger.Infof("Fetching access token for %s", cspCmdConfig.ServiceAccountEmail)
	accessToken, err := gcpclient.GenerateTemporaryAccessToken(cspCmdConfig.ServiceAccountEmail, cspCmdConfig.Reason)
	if err != nil {
		return err
	}

	util.Logger.Infof("Running: [cloud_sql_proxy %s]\n\n", strings.Join(cloudSQLProxyCmdArgs, " "))
	cloudSQLProxyAuth := append(cloudSQLProxyCmdArgs, "-token", accessToken.GetAccessToken())
	c := exec.Command(cloudSQLProxyPath, cloudSQLProxyAuth...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	if err := c.Run(); err != nil {
		fullCmd := fmt.Sprintf("cloud_sql_proxy %s", strings.Join(cloudSQLProxyCmdArgs, " "))
		return errorsutil.New(fmt.Sprintf("Failed to run command [%s]", fullCmd), err)
	}
	return nil
}
