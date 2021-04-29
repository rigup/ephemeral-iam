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
	"github.com/lithammer/dedent"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	"github.com/rigup/ephemeral-iam/internal/gcpclient"
	"github.com/rigup/ephemeral-iam/internal/proxy"
	"github.com/rigup/ephemeral-iam/pkg/options"
)

var apCmdConfig options.CmdConfig

func newCmdAssumePrivileges() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "assume-privileges",
		Aliases: []string{"priv"},
		Short:   "Configure gcloud to make API calls as the provided service account [alias: priv]",
		Long: dedent.Dedent(`
			The "assume-privileges" command fetches short-lived credentials for the provided service Account
			and configures gcloud to proxy its traffic through an auth proxy. This auth proxy sets the
			authorization header to the OAuth2 token generated for the provided service account. Once
			the credentials have expired, the auth proxy is shut down and the gcloud config is restored.
			
			The reason flag is used to add additional metadata to audit logs.  The provided reason will
			be in 'protoPayload.requestMetadata.requestAttributes.reason'.`),
		Example: dedent.Dedent(`
				eiam assume-privileges \
				  --service-account-email example@my-project.iam.gserviceaccount.com \
				  --reason "Emergency security patch (JIRA-1234)"`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().VisitAll(options.CheckRequired)

			if err := util.FormatReason(&apCmdConfig.Reason); err != nil {
				return err
			}

			if !options.YesOption {
				util.Confirm(map[string]string{
					"Project":         apCmdConfig.Project,
					"Service Account": apCmdConfig.ServiceAccountEmail,
					"Reason":          apCmdConfig.Reason,
				})
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return startPrivilegedSession()
		},
	}

	options.AddServiceAccountEmailFlag(cmd.Flags(), &apCmdConfig.ServiceAccountEmail, true)
	options.AddReasonFlag(cmd.Flags(), &apCmdConfig.Reason, true)
	options.AddProjectFlag(cmd.Flags(), &apCmdConfig.Project)

	return cmd
}

func startPrivilegedSession() error {
	hasAccess, err := gcpclient.CanImpersonate(
		apCmdConfig.Project,
		apCmdConfig.ServiceAccountEmail,
		apCmdConfig.Reason,
	)
	if err != nil {
		return err
	} else if !hasAccess {
		util.Logger.Fatalln("You do not have access to impersonate this service account")
	}

	util.Logger.Info("Fetching short-lived access token for ", apCmdConfig.ServiceAccountEmail)
	accessToken, err := gcpclient.GenerateTemporaryAccessToken(apCmdConfig.ServiceAccountEmail, apCmdConfig.Reason)
	if err != nil {
		return err
	}

	util.Logger.Info("Configuring gcloud to use auth proxy")
	if gcpclient.ConfigureGcloudProxy(apCmdConfig.Project) != nil {
		return err
	}

	clusters, err := gcpclient.GetClusters(apCmdConfig.Project, apCmdConfig.Reason)
	if err != nil {
		return err
	}

	defaultCluster := map[string]string{}
	if len(clusters) == 0 {
		util.Logger.Warnf("No clusters found in %s", apCmdConfig.Project)
	} else if len(clusters) == 1 {
		defaultCluster = clusters[0]
	} else {
		clusterNames := []string{}
		for _, cl := range clusters {
			clusterNames = append(clusterNames, cl["name"])
		}
		prompt := promptui.Select{
			Label: "Select the default cluster to use",
			Items: clusterNames,
		}

		_, result, err := prompt.Run()
		if err != nil {
			util.Logger.Warn("No cluster default cluster will be configured")
		} else {
			for _, cl := range clusters {
				if cl["name"] == result {
					defaultCluster = cl
					break
				}
			}
			if defaultCluster == nil {
				util.Logger.Warnf("Invalid cluster name selected: %s", result)
			}
		}
	}
	token := accessToken.GetAccessToken()
	expirationDate := accessToken.GetExpireTime().AsTime()
	return proxy.StartProxyServer(
		token,
		apCmdConfig.Reason,
		apCmdConfig.ServiceAccountEmail,
		apCmdConfig.Project,
		expirationDate,
		defaultCluster,
	)
}
