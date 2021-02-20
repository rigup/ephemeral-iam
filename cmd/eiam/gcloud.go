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

var gcloudCmdArgs []string

var runGcloudCmd = &cobra.Command{
	Use:                "gcloud [GCLOUD_ARGS]",
	Short:              "Run a gcloud command with the permissions of the specified service account",
	Long:               `TODO`,
	Args:               cobra.ArbitraryArgs,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	PreRun: func(cmd *cobra.Command, args []string) {
		gcloudCmdArgs = extractUnknownArgs(cmd.Flags(), os.Args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		project, err := gcpclient.GetCurrentProject()
		handleErr(err)

		fmt.Println()
		logger.Infof("Project:            %s\n", project)
		logger.Infof("Service Account:    %s\n", serviceAccountEmail)
		logger.Infof("Reason:             %s\n", reason)
		logger.Infof("Command:            gcloud %s\n\n", strings.Join(gcloudCmdArgs, " "))

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

		// gcloud reads the CLOUDSDK_CORE_REQUEST_REASON environment variable
		// and sets the X-Goog-Request-Reason header in API requests to its value
		reasonHeader := fmt.Sprintf("CLOUDSDK_CORE_REQUEST_REASON=%s", reason)

		// There has to be a better way to do this...
		logger.Infof("Running: [gcloud %s]\n\n", strings.Join(gcloudCmdArgs, " "))
		gcloudCmdArgs = append(gcloudCmdArgs, "--impersonate-service-account", serviceAccountEmail, "--verbosity=error")
		c := exec.Command("gcloud", gcloudCmdArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Env = append(os.Environ(), reasonHeader)

		if err := c.Run(); err != nil {
			fullCmd := fmt.Sprintf("gcloud %s", strings.Join(gcloudCmdArgs, " "))
			logger.Errorf("Error: %v for command [%s]", err, fullCmd)
		}
	},
}

func init() {
	rootCmd.AddCommand(runGcloudCmd)
	runGcloudCmd.Flags().StringVarP(&serviceAccountEmail, "serviceAccountEmail", "s", "", "The email address for the service account to impersonate (required)")
	runGcloudCmd.Flags().StringVarP(&reason, "reason", "r", "", "A detailed rationale for assuming higher permissions (required)")
	runGcloudCmd.MarkFlagRequired("serviceAccountEmail")
	runGcloudCmd.MarkFlagRequired("reason")

	// if len(os.Args) > 2 && os.Args[1] == "gcloud" {
	// 	gcloudCmdArgs = extractUnknownArgs(runGcloudCmd.Flags(), os.Args)
	// }
}
