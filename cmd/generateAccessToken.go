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
	"context"
	"fmt"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	gcp "google.golang.org/api/container/v1"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"

	"github.com/jessesomerville/gcp-iam-escalate/google"
	"github.com/jessesomerville/gcp-iam-escalate/loghandler"
)

var logger *logrus.Logger
var credentialsClient *credentials.IamCredentialsClient

var serviceAccountEmail string

// generateAccessTokenCmd represents the generateAccessToken command
var generateAccessTokenCmd = &cobra.Command{
	Use:   "generateAccessToken",
	Short: "Generate a short-lived OAuth 2 token for a given service account",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Fetching short-lived access token for: ", serviceAccountEmail)
		accessToken, err := generateTemporaryAccessToken()
		emperror.Panic(err)
		createPrivilegedClient(accessToken)
	},
}

func init() {
	logger = loghandler.GetLogger()
	credentialsClient = google.GetGCPClient()
	rootCmd.AddCommand(generateAccessTokenCmd)
	generateAccessTokenCmd.Flags().StringVar(&serviceAccountEmail, "serviceAccountEmail", "", "The email address for the service account to impersonate")
	generateAccessTokenCmd.MarkFlagRequired("serviceAccountEmail")
}

func generateTemporaryAccessToken() (*credentialspb.GenerateAccessTokenResponse, error) {

	sessionDuration := &duration.Duration{
		Seconds: 600, // Expire after 10 minutes
	}

	req := credentialspb.GenerateAccessTokenRequest{
		Name:     fmt.Sprintf("projects/-/serviceAccounts/%s", serviceAccountEmail),
		Lifetime: sessionDuration,
		Scope: []string{
			gcp.CloudPlatformScope,
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}

	ctx := context.Background()
	resp, err := credentialsClient.GenerateAccessToken(ctx, &req)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "Failed to generate GCP access token for service account", "service account", serviceAccountEmail)
	}
	return resp, nil
}

func createPrivilegedClient(accessToken *credentialspb.GenerateAccessTokenResponse) {
	token := &oauth2.Token{
		AccessToken: accessToken.GetAccessToken(),
	}
	tokenSource := oauth2.StaticTokenSource(token)
	google.GetPrivilegedClient(tokenSource)
	logger.Info("Created new API client as impersonated service account")

	/*
		I am trying to think of the best way to drop users into the privileged session.
		One way would be to manually write the OAuth token to `~/.config/gcloud/access_tokens.db`
		then set their current gcloud configuration to that.

		Another possible solution -- albiet, an over engineered one -- would be to set up a proxy
		that replaces the auth in API calls with the OAuth token.  `gcloud` can be configured to
		use a proxy by setting `gcloud config set proxy/...`.
	*/
}
