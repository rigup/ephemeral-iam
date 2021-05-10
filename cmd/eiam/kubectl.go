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
	kubectlCmdArgs   []string
	kubectlCmdConfig options.CmdConfig
)

func NewCmdKubectl() *cobra.Command {
	cmd := WrapperCommand("kubectl", &kubectlCmdArgs, &kubectlCmdConfig, runKubectlCommand)

	cmd.Example = dedent.Dedent(`
		eiam kubectl pods -o json \
		--service-account-email example@my-project.iam.gserviceaccount.com \
		--reason "Debugging for (JIRA-1234)"
			
		eiam kubectl pods -o json \
		-s example@my-project.iam.gserviceaccount.com -r "example" \
		| jq`)
	return cmd
}

func runKubectlCommand() error {
	hasAccess, err := gcpclient.CanImpersonate(kubectlCmdConfig.Project, kubectlCmdConfig.ServiceAccountEmail)
	if err != nil {
		return err
	} else if !hasAccess {
		util.Logger.Fatalln("You do not have access to impersonate this service account")
	}

	util.Logger.Infof("Fetching access token for %s", kubectlCmdConfig.ServiceAccountEmail)
	accessToken, err := gcpclient.GenerateTemporaryAccessToken(
		kubectlCmdConfig.ServiceAccountEmail,
		kubectlCmdConfig.Reason,
	)
	if err != nil {
		return err
	}

	util.Logger.Infof("Running: [kubectl %s]\n\n", strings.Join(kubectlCmdArgs, " "))
	kubectlAuth := append(kubectlCmdArgs, "--token", accessToken.GetAccessToken())
	kubectl := viper.GetString("binarypaths.kubectl")
	c := exec.Command(kubectl, kubectlAuth...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	if err := c.Run(); err != nil {
		fullCmd := fmt.Sprintf("kubectl %s", strings.Join(kubectlCmdArgs, " "))
		return errorsutil.New(fmt.Sprintf("Failed to run command [%s]", fullCmd), err)
	}

	return nil
}
