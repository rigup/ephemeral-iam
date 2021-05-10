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

package eiam

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	"github.com/rigup/ephemeral-iam/pkg/options"
)

// WrapperCommand constructs an eiam command that serves as a wrapper around another
// external command (e.g. gcloud, kubectl, cloud_sql_proxy).
func WrapperCommand(
	cmdName string,
	cmdArgs *[]string,
	cmdOpts *options.CmdConfig,
	cmdRun func() error,
) *cobra.Command {
	short := fmt.Sprintf("Run a %s command with the permissions of the specified service account", cmdName)

	cmd := &cobra.Command{
		Use:                fmt.Sprintf("%s [GCLOUD_ARGS]", cmdName),
		Short:              short,
		Args:               cobra.ArbitraryArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().VisitAll(options.CheckRequired)

			argsToParse := os.Args
			if flag.Lookup("test.v") != nil {
				// If we are running tests, we need to parse args as well.
				argsToParse = append(argsToParse, args...)

				// We also need to clear any args from previous commands ran by tests.
				if len(*cmdArgs) != 0 {
					*cmdArgs = []string{}
				}
			}

			*cmdArgs = util.ExtractUnknownArgs(cmd.Flags(), argsToParse)
			if err := util.FormatReason(&cmdOpts.Reason); err != nil {
				return err
			}

			if !options.YesOption {
				util.Confirm(map[string]string{
					"Project":         cmdOpts.Project,
					"Service Account": cmdOpts.ServiceAccountEmail,
					"Reason":          cmdOpts.Reason,
					"Command":         fmt.Sprintf("%s %s", cmdName, strings.Join(*cmdArgs, " ")),
				})
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdRun()
		},
	}

	options.AddServiceAccountEmailFlag(cmd.Flags(), &cmdOpts.ServiceAccountEmail, true)
	options.AddReasonFlag(cmd.Flags(), &cmdOpts.Reason, true)
	options.AddProjectFlag(cmd.Flags(), &cmdOpts.Project)

	return cmd
}
