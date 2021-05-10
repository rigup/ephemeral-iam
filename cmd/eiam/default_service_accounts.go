package eiam

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/iam/v1"

	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
	"github.com/rigup/ephemeral-iam/internal/gcpclient"
	"github.com/rigup/ephemeral-iam/pkg/options"
)

var project string

func NewCmdDefaultServiceAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "default-service-accounts",
		Aliases: []string{"default-sa"},
		Short:   "Configure default service accounts to use in other commands [alias: default-sa]",
	}
	cmd.AddCommand(NewCmdSetDefaultServiceAccount())
	cmd.AddCommand(NewCmdListDefaultServiceAccounts())
	return cmd
}

func NewCmdSetDefaultServiceAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set a default privileged service account to impersonate for a given GCP project",
		RunE: func(cmd *cobra.Command, args []string) error {
			availableSAs, err := gcpclient.FetchAvailableServiceAccounts(project)
			if err != nil {
				return err
			}
			if len(availableSAs) == 0 {
				util.Logger.Warnf("You cannot impersonate any service accounts in %s", project)
				return nil
			}

			selected, err := selectServiceAccount(availableSAs)
			if err != nil {
				return errorsutil.New("Failed to get selected service account: %v", err)
			}

			defaultSAs := viper.GetStringMapString(appconfig.DefaultServiceAccounts)
			defaultSAs[project] = selected
			viper.Set(appconfig.DefaultServiceAccounts, defaultSAs)
			if err := viper.WriteConfig(); err != nil {
				return errorsutil.New("Failed to write updated configuration", err)
			}

			util.Logger.Infof("Set default service account for %s to %s", project, selected)
			return nil
		},
	}
	options.AddProjectFlag(cmd.Flags(), &project)
	return cmd
}

func NewCmdListDefaultServiceAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured default service accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			defaultSAs := viper.GetStringMapString(appconfig.DefaultServiceAccounts)
			if len(defaultSAs) == 0 {
				util.Logger.Warn("You have not set any default service accounts")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
			fmt.Fprintln(w, "\nPROJECT\tSERVICE ACCOUNT")
			for proj, sa := range defaultSAs {
				fmt.Fprintf(w, "%s\t%s\n", proj, sa)
			}
			w.Flush()
			fmt.Println()
			return nil
		},
	}
	return cmd
}

func selectServiceAccount(availableSAs []*iam.ServiceAccount) (string, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   " ►  {{ .Email | blue }}",
		Inactive: "  {{ .Email | cyan }}",
		Selected: " ►  {{ .Email | green }}",
	}

	prompt := promptui.Select{
		Label:        "Select Service Account",
		Items:        availableSAs,
		Templates:    templates,
		HideSelected: true,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return availableSAs[i].Email, nil
}
