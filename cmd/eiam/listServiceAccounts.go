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
package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"
	"google.golang.org/api/iam/v1"

	"github.com/jessesomerville/ephemeral-iam/internal/gcpclient"
)

// listServiceAccountsCmd represents the listServiceAccounts command
var listServiceAccountsCmd = &cobra.Command{
	Use:   "listServiceAccounts",
	Short: "List service accounts that can be impersonated",
	Long: `
	The "listServiceAccounts" command fetches all Cloud IAM Service Accounts in the current
	GCP project (as determined by the activated gcloud config) and checks each of them to see
	which ones the current user has access to impersonate.

	NOTE: For this to work properly, the current user must have access to list service accounts
	in the current project.

	Example:
	  gcp-iam-elevate listServiceAccounts`,
	Run: func(cmd *cobra.Command, args []string) {
		err := fetchAvailableServiceAccounts()
		emperror.Panic(err)
	},
}

func init() {
	rootCmd.AddCommand(listServiceAccountsCmd)
}

func fetchAvailableServiceAccounts() error {
	project, err := gcpclient.GetCurrentProject()
	if err != nil {
		return errors.WrapIf(err, "Failed to get current GCP project from gcloud config")
	}
	logger.Info("Using current project: ", project)

	serviceAccounts, err := gcpclient.GetServiceAccounts(project)
	if err != nil {
		return err
	}

	var availableSAs []*iam.ServiceAccount
	for _, serviceAccount := range serviceAccounts {
		hasAccess, err := gcpclient.CanImpersonate(project, serviceAccount.Email)
		if err != nil {
			return errors.WrapIf(err, "Error checking IAM permissions")
		} else if hasAccess {
			availableSAs = append(availableSAs, serviceAccount)
		}
	}
	if len(availableSAs) == 0 {
		logger.Warning("You do not have access to impersonate any accounts in this project")
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
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", sa.Email, desc[0]))
		} else {
			firstLine, remaining := desc[0], desc[1:]
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s", sa.Email, firstLine))
			for _, line := range remaining {
				fmt.Fprintln(w, fmt.Sprintf("%s\t%s", " ", line))
			}
		}
	}
	w.Flush()
}
