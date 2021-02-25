package cmd

import (
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/cmd/options"
	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/gcpclient"
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/proxy"
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
			be in 'protoPatload.requestMetadata.requestAttributes.reason'.`),
		Example: dedent.Dedent(`
				eiam assume-privileges \
				  --serviceAccountEmail example@my-project.iam.gserviceaccount.com \
				  --reason "Emergency security patch (JIRA-1234)"`),
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(options.CheckRequired)

			util.CheckError(util.FormatReason(&apCmdConfig.Reason))

			if !options.YesOption {
				util.Confirm(map[string]string{
					"Project":         apCmdConfig.Project,
					"Service Account": apCmdConfig.ServiceAccountEmail,
					"Reason":          apCmdConfig.Reason,
				})
			}
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
	if err := gcpclient.ConfigureGcloudProxy(); err != nil {
		return err
	}

	return proxy.StartProxyServer(accessToken, apCmdConfig.Reason, apCmdConfig.ServiceAccountEmail)
}
