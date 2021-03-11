package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/lithammer/dedent"
	"github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"
	"google.golang.org/api/iam/v1"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/cmd/options"
	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/gcpclient"
)

var listCmdConfig options.CmdConfig

func newCmdListServiceAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-service-accounts",
		Aliases: []string{"list"},
		Short:   "List service accounts that can be impersonated [alias: list]",
		Long: dedent.Dedent(`
			The "list-service-accounts" command fetches all Cloud IAM Service Accounts in the current
			GCP project (as determined by the activated gcloud config) and checks each of them to see
			which ones the current user has access to impersonate.`),
		Example: dedent.Dedent(`
			eiam list-service-accounts
		
			eiam list`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchAvailableServiceAccounts()
		},
	}
	options.AddProjectFlag(cmd.Flags(), &listCmdConfig.Project)

	return cmd
}

func fetchAvailableServiceAccounts() error {
	util.Logger.Infof("Using current project: %s", listCmdConfig.Project)

	serviceAccounts, err := gcpclient.GetServiceAccounts(listCmdConfig.Project, listCmdConfig.Reason)
	if err != nil {
		return err
	}

	var availableSAs []*iam.ServiceAccount
	for _, serviceAccount := range serviceAccounts {
		hasAccess, err := gcpclient.CanImpersonate(listCmdConfig.Project, serviceAccount.Email, listCmdConfig.Reason)
		if err != nil {
			return fmt.Errorf("Error checking IAM permissions: %v", err)
		} else if hasAccess {
			availableSAs = append(availableSAs, serviceAccount)
		}
	}
	if len(availableSAs) == 0 {
		util.Logger.Warning("You do not have access to impersonate any accounts in this project")
	}

	printColumns(availableSAs)
	return nil
}

func printColumns(serviceAccounts []*iam.ServiceAccount) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
	fmt.Println()
	fmt.Fprintln(w, "EMAIL\tDESCRIPTION")
	for _, sa := range serviceAccounts {
		desc := strings.Split(wordwrap.WrapString(sa.Description, 75), "\n")
		if len(desc) == 1 {
			fmt.Fprintf(w, "%s\t%s\n", sa.Email, desc[0])
		} else {
			firstLine, remaining := desc[0], desc[1:]
			fmt.Fprintf(w, "%s\t%s\n", sa.Email, firstLine)
			for _, line := range remaining {
				fmt.Fprintf(w, "%s\t%s\n", " ", line)
			}
		}
	}
	w.Flush()
}
