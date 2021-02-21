/*
Copyright © 2021 Jesse Somerville

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jessesomerville/ephemeral-iam/internal/gcpclient"
)

var (
	kubectlCmdArgs []string
	zone           string
)

var runKubectlCmd = &cobra.Command{
	Use:   "kubectl [KUBECTL_ARGS]",
	Short: "Run a kubectl command with the permissions of the specified service account",
	Long: `
The "kubectl" command runs the provided kubectl command with the permissions of the specified
service account. Output from the kubectl command is able to be piped into other commands.

Examples:
	eiam kubectl pods -o json \
	  --serviceAccountEmail example@my-project.iam.gserviceaccount.com \
	  --reason "Debugging for (JIRA-1234)"
		
	eiam kubectl pods -o json \
	  -s example@my-project.iam.gserviceaccount.com -r "example" \
	  | jq`,
	Args:               cobra.ArbitraryArgs,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	PreRun: func(cmd *cobra.Command, args []string) {
		kubectlCmdArgs = extractUnknownArgs(cmd.Flags(), os.Args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		project, err := gcpclient.GetCurrentProject()
		handleErr(err)

		fmt.Println()
		logger.Infof("Project:            %s\n", project)
		logger.Infof("Service Account:    %s\n", serviceAccountEmail)
		logger.Infof("Reason:             %s\n", reason)
		logger.Infof("Command:            kubectl %s\n\n", strings.Join(kubectlCmdArgs, " "))

		if !Accept {
			if err := confirm(); err != nil {
				os.Exit(0)
			}
		}

		reason, err := formatReason(reason)
		handleErr(err)

		logger.Infof("Checking sufficient permissions to impersonate %s", serviceAccountEmail)

		hasAccess, err := gcpclient.CanImpersonate(project, serviceAccountEmail, reason)
		handleErr(err)
		if !hasAccess {
			logger.Error("You do not have access to impersonate this service account")
			os.Exit(1)
		}

		logger.Info("Fetching access token for ", serviceAccountEmail)

		gcpClientWithReason, err := gcpclient.WithReason(reason)
		handleErr(err)

		accessToken, err := gcpclient.GenerateTemporaryAccessToken(serviceAccountEmail, gcpClientWithReason)
		handleErr(err)

		logger.Infof("Running: [kubectl %s]\n\n", strings.Join(kubectlCmdArgs, " "))
		kubectlAuth := append(kubectlCmdArgs, "--token", accessToken.GetAccessToken())
		c := exec.Command("kubectl", kubectlAuth...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			fullCmd := fmt.Sprintf("kubectl %s", strings.Join(kubectlCmdArgs, " "))
			logger.Errorf("%v for command [kubectl %s]", err, fullCmd)
		}

	},
}

func init() {
	rootCmd.AddCommand(runKubectlCmd)
	runKubectlCmd.Flags().StringVarP(&serviceAccountEmail, "serviceAccountEmail", "s", "", "The email address for the service account to impersonate (required)")
	runKubectlCmd.Flags().StringVarP(&reason, "reason", "r", "", "A detailed rationale for assuming higher permissions (required)")
	runKubectlCmd.MarkFlagRequired("serviceAccountEmail")
	runKubectlCmd.MarkFlagRequired("reason")
}
