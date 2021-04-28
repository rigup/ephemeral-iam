package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/lithammer/dedent"
	"github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"
	"google.golang.org/api/iam/v1"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	"github.com/rigup/ephemeral-iam/internal/gcpclient"
	"github.com/rigup/ephemeral-iam/pkg/options"
)

var (
	listCmdConfig options.CmdConfig
	wg            sync.WaitGroup
)

func newCmdListServiceAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "list-service-accounts",
		Aliases:    []string{"list"},
		Short:      "List service accounts that can be impersonated [alias: list]",
		SuggestFor: []string{"ls"},
		Long: dedent.Dedent(`
			The "list-service-accounts" command fetches all Cloud IAM Service Accounts in the current
			GCP project (as determined by the activated gcloud config) and checks each of them to see
			which ones the current user has access to impersonate.`),
		Example: dedent.Dedent(`
			$ eiam list-service-accounts
			$ eiam list`),
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(options.CheckRequired)
		},
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
	util.Logger.Infof("Checking %d service accounts in %s", len(serviceAccounts), listCmdConfig.Project)

	wg.Add(len(serviceAccounts))

	var availableSAs []*iam.ServiceAccount
	for _, svcAcct := range serviceAccounts {
		go func(serviceAccount *iam.ServiceAccount) {
			hasAccess, err := gcpclient.CanImpersonate(listCmdConfig.Project, serviceAccount.Email, listCmdConfig.Reason)
			if err != nil {
				util.Logger.Errorf("error checking IAM permissions: %v", err)
			} else if hasAccess {
				availableSAs = append(availableSAs, serviceAccount)
			}
			wg.Done()
		}(svcAcct)
	}
	wg.Wait()

	if len(availableSAs) == 0 {
		util.Logger.Warning("You do not have access to impersonate any accounts in this project")
		return nil
	}

	printColumns(availableSAs)
	return nil
}

func printColumns(serviceAccounts []*iam.ServiceAccount) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
	fmt.Fprintln(w, "\nEMAIL\tDESCRIPTION")
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
