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

package cmd

import (
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
	gcloudCmdArgs   []string
	gcloudCmdConfig options.CmdConfig
)

func newCmdGcloud() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcloud [GCLOUD_ARGS]",
		Short: "Run a gcloud command with the permissions of the specified service account",
		Long: dedent.Dedent(`
			The "gcloud" command runs the provided gcloud command with the permissions of the specified
			service account. Output from the gcloud command is able to be piped into other commands.`),
		Example: dedent.Dedent(`
			eiam gcloud compute instances list --format=json \
			--service-account-email example@my-project.iam.gserviceaccount.com \
			--reason "Debugging for (JIRA-1234)"
			
			eiam gcloud compute instances list --format=json \
			-s example@my-project.iam.gserviceaccount.com -r "example" \
			| jq`),
		Args:               cobra.ArbitraryArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().VisitAll(options.CheckRequired)

			gcloudCmdArgs = util.ExtractUnknownArgs(cmd.Flags(), os.Args)
			if err := util.FormatReason(&gcloudCmdConfig.Reason); err != nil {
				return err
			}

			if !options.YesOption {
				util.Confirm(map[string]string{
					"Project":         gcloudCmdConfig.Project,
					"Service Account": gcloudCmdConfig.ServiceAccountEmail,
					"Reason":          gcloudCmdConfig.Reason,
					"Command":         fmt.Sprintf("gcloud %s", strings.Join(gcloudCmdArgs, " ")),
				})
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGcloudCommand()
		},
	}

	options.AddServiceAccountEmailFlag(cmd.Flags(), &gcloudCmdConfig.ServiceAccountEmail, true)
	options.AddReasonFlag(cmd.Flags(), &gcloudCmdConfig.Reason, true)
	options.AddProjectFlag(cmd.Flags(), &gcloudCmdConfig.Project)

	return cmd
}

func runGcloudCommand() error {
	hasAccess, err := gcpclient.CanImpersonate(
		gcloudCmdConfig.Project,
		gcloudCmdConfig.ServiceAccountEmail,
		gcloudCmdConfig.Reason,
	)
	if err != nil {
		return err
	} else if !hasAccess {
		util.Logger.Fatalln("You do not have access to impersonate this service account")
	}

	// gcloud reads the CLOUDSDK_CORE_REQUEST_REASON environment variable
	// and sets the X-Goog-Request-Reason header in API requests to its value.
	reasonHeader := fmt.Sprintf("CLOUDSDK_CORE_REQUEST_REASON=%s", gcloudCmdConfig.Reason)

	// There has to be a better way to do this...
	util.Logger.Infof("Running: [gcloud %s]\n\n", strings.Join(gcloudCmdArgs, " "))
	gcloudCmdArgs = append(
		gcloudCmdArgs,
		"--impersonate-service-account", gcloudCmdConfig.ServiceAccountEmail,
		"--verbosity=error",
	)
	gcloud := viper.GetString("binarypaths.gcloud")
	c := exec.Command(gcloud, gcloudCmdArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(os.Environ(), reasonHeader)

	if err := c.Run(); err != nil {
		fullCmd := fmt.Sprintf("gcloud %s", strings.Join(gcloudCmdArgs, " "))
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: fmt.Sprintf("Failed to run command [%s]", fullCmd),
			Err: err,
		}
	}
	return nil
}
