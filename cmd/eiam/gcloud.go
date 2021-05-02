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

func NewCmdGcloud() *cobra.Command {
	cmd := WrapperCommand("gcloud", &gcloudCmdArgs, &gcloudCmdConfig, runGcloudCommand)
	cmd.Example = dedent.Dedent(`
		eiam gcloud compute instances list --format=json \
		--service-account-email example@my-project.iam.gserviceaccount.com \
		--reason "Debugging for (JIRA-1234)"
		
		eiam gcloud compute instances list --format=json \
		-s example@my-project.iam.gserviceaccount.com -r "example" \
		| jq`)
	return cmd
}

func runGcloudCommand() error {
	hasAccess, err := gcpclient.CanImpersonate(gcloudCmdConfig.Project, gcloudCmdConfig.ServiceAccountEmail)
	if err != nil {
		return err
	} else if !hasAccess {
		err = fmt.Errorf("cannot impersonate %s", gcloudCmdConfig.ServiceAccountEmail)
		sErr := errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: "You do not have access to impersonate this service account",
			Err: err,
		}
		fmt.Printf("\n\nRIGHT HERE: %+v\n\n", sErr)
		return sErr
	}

	// gcloud reads the CLOUDSDK_CORE_REQUEST_REASON environment variable
	// and sets the X-Goog-Request-Reason header in API requests to its value.
	reasonHeader := fmt.Sprintf("CLOUDSDK_CORE_REQUEST_REASON=%s", gcloudCmdConfig.Reason)

	// There has to be a better way to do this...
	util.Logger.Infof("Running: [gcloud %s]\n\n", strings.Join(gcloudCmdArgs, " "))
	cmdArgs := append(
		gcloudCmdArgs,
		"--impersonate-service-account", gcloudCmdConfig.ServiceAccountEmail,
		"--verbosity=error",
	)
	gcloud := viper.GetString("binarypaths.gcloud")
	c := exec.Command(gcloud, cmdArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Env = append(os.Environ(), reasonHeader)

	if err := c.Run(); err != nil {
		fullCmd := fmt.Sprintf("gcloud %s", strings.Join(gcloudCmdArgs, " "))
		return errorsutil.New(fmt.Sprintf("Failed to run command [%s]", fullCmd), err)
	}
	return nil
}
