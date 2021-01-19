/*
Copyright © 2021 Jesse Somerville <jssomerville2@gmail.com>

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
package cmd

import (
	"os"

	"emperror.dev/emperror"
	"github.com/spf13/cobra"

	"github.com/jessesomerville/gcp-iam-escalate/gcpclient"
	"github.com/jessesomerville/gcp-iam-escalate/proxy"
)

// TODO: include the `reason` flag in Slack Alerts
var serviceAccountEmail string
var reason string

// generateAccessTokenCmd represents the generateAccessToken command
var assumePrivilegesCmd = &cobra.Command{
	Use:   "assumePrivileges",
	Short: "Configure gcloud to make API calls as the provided service account",
	Long: `
	The "assumePrivileges" command fetches short-lived credentials for the provided service Account
	and configures gcloud to proxy its traffic through an auth proxy. This auth proxy sets the
	authorization header to the OAuth2 token generated for the provided service account. Once
	the credentials have expired, the auth proxy is shut down and the gcloud config is restored.

	Example:
	  gcp_iam_escalate assumePrivileges --serviceAccountEmail example@my-project.iam.gserviceaccount.com`,
	Run: func(cmd *cobra.Command, args []string) {
		project, err := gcpclient.GetCurrentProject()
		emperror.Panic(err)
		hasAccess, err := gcpclient.CanImpersonate(project, serviceAccountEmail)
		emperror.Panic(err)
		if !hasAccess {
			logger.Error("You do not have access to impersonate this service account")
			os.Exit(1)
		}

		logger.Info("Fetching short-lived access token for ", serviceAccountEmail)
		accessToken, err := gcpclient.GenerateTemporaryAccessToken(serviceAccountEmail)
		emperror.Panic(err)

		logger.Info("Configuring gcloud to use auth proxy")
		emperror.Panic(gcpclient.ConfigureGcloudProxy())

		proxy.StartProxyServer(accessToken)
	},
}

func init() {
	rootCmd.AddCommand(assumePrivilegesCmd)
	assumePrivilegesCmd.Flags().StringVar(&serviceAccountEmail, "serviceAccountEmail", "", "The email address for the service account to impersonate (required)")
	assumePrivilegesCmd.Flags().StringVar(&reason, "reason", "", "A detailed rationale for assuming higher permissions (required)")
	assumePrivilegesCmd.MarkFlagRequired("serviceAccountEmail")
	assumePrivilegesCmd.MarkFlagRequired("reason")
}
