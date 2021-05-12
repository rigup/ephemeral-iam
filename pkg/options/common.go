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

package options

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
	"github.com/rigup/ephemeral-iam/internal/gcpclient"
)

// Flag annotation strings.
const (
	RequiredAnnotation = "eiam_required_flag"
)

// YesOption designates whether to prompt for confirmation or not.
var (
	YesOption = false
)

// Flag names and shorthands.
var (
	// FormatFlag controls the output format for a command.
	FormatFlag = flagName{"format", "f"}

	// ProjectFlag sets the GCP project to use for a command.
	ProjectFlag = flagName{"project", "p"}

	// ReasonFlag enforces that a rationale be given for a command.
	ReasonFlag = flagName{"reason", "R"}

	// RegionFlag sets the GCP region to use for a command.
	RegionFlag = flagName{"region", "r"}

	// ServiceAccountEmailFlag sets the service account to use for a command.
	ServiceAccountEmailFlag = flagName{"service-account-email", "s"}

	// YesFlag is a boolean that when set to true ignores non-required user prompts.
	YesFlag = flagName{"yes", "y"}

	// ZoneFlag sets the GCP zone to use for a command.
	ZoneFlag = flagName{"zone", "z"}
)

type flagName struct {
	Name      string
	Shorthand string
}

// CmdConfig holds the values passed to a command.
type CmdConfig struct {
	ComputeInstance     string
	Project             string
	PubSubTopic         string
	Reason              string
	Region              string
	ServiceAccountEmail string
	StorageBucket       string
	Zone                string
}

// AddPersistentFlags add persistent flags to the root command.
func AddPersistentFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&YesOption, YesFlag.Name, YesFlag.Shorthand, YesOption, "Assume 'yes' to all prompts")

	currLogFmt := viper.GetString(appconfig.LoggingFormat)
	fs.StringP(FormatFlag.Name, FormatFlag.Shorthand, currLogFmt, "Set the output of the current command")
	if err := viper.BindPFlag(appconfig.LoggingFormat, fs.Lookup(FormatFlag.Name)); err != nil {
		util.Logger.Fatalf("failed to add `--format` flag to root command")
	}
}

// AddProjectFlag adds the --project/-p flag to the command.
func AddProjectFlag(fs *pflag.FlagSet, project *string, required bool) {
	defaultVal, err := gcpclient.GetCurrentProject()
	errorsutil.CheckError(err)

	fs.StringVarP(
		project,
		ProjectFlag.Name,
		ProjectFlag.Shorthand,
		defaultVal,
		"The GCP project. Inherits from the active gcloud config by default",
	)
	if defaultVal == "" || required {
		if err := fs.SetAnnotation(ProjectFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}

// AddRegionFlag adds the --region/-r flag to the command.
func AddRegionFlag(fs *pflag.FlagSet, region *string, required bool) {
	defaultVal, err := gcpclient.GetCurrentRegion()
	errorsutil.CheckError(err)

	fs.StringVarP(
		region,
		RegionFlag.Name,
		RegionFlag.Shorthand,
		defaultVal,
		"The GCP region. Inherits from the active gcloud config by default",
	)
	if required {
		if err := fs.SetAnnotation(RegionFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}

// AddZoneFlag adds the --zone/-z flag to the command.
func AddZoneFlag(fs *pflag.FlagSet, zone *string, required bool) {
	defaultVal, err := gcpclient.GetCurrentZone()
	errorsutil.CheckError(err)

	fs.StringVarP(
		zone,
		ZoneFlag.Name,
		ZoneFlag.Shorthand,
		defaultVal,
		"The GCP zone. Inherits from the active gcloud config by default",
	)
	if required {
		if err := fs.SetAnnotation(ZoneFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}

// AddServiceAccountEmailFlag adds the --service-account-email/-s flag.
func AddServiceAccountEmailFlag(fs *pflag.FlagSet, serviceAccountEmail *string, required bool) {
	defaultVal := ""
	defaultSAs := viper.GetStringMapString(appconfig.DefaultServiceAccounts)
	activeProject, err := gcpclient.GetCurrentProject()
	errorsutil.CheckError(err)

	if val, ok := defaultSAs[activeProject]; ok {
		defaultVal = val
	}
	fs.StringVarP(
		serviceAccountEmail,
		ServiceAccountEmailFlag.Name,
		ServiceAccountEmailFlag.Shorthand,
		defaultVal,
		"The email address for the service account. Defaults to the configured default account for the current project",
	)
	if required {
		if err := fs.SetAnnotation(ServiceAccountEmailFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}

// AddReasonFlag adds the --reason/-R flag.
func AddReasonFlag(fs *pflag.FlagSet, reason *string, required bool) {
	fs.StringVarP(
		reason,
		ReasonFlag.Name,
		ReasonFlag.Shorthand,
		"",
		"A detailed rationale for assuming higher permissions",
	)
	if required {
		if err := fs.SetAnnotation(ReasonFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}

// CheckRequired ensures that a command's required flags have been set. The only
// way to iterate over every flag in a pflag.FlagSet is with the VisitAll command.
// VisitAll takes a function as a parameter and calls that function on each flag in the
// flag set. This input function does not allow you to return errors; therefore,
// CheckRequired creates an error channel for it to send any errors that it encounters
// to then returns them.
func CheckRequired(flags *pflag.FlagSet) error {
	for err := range visitAllFlags(flags) {
		if err != nil {
			return err
		}
	}

	return nil
}

func visitAllFlags(flags *pflag.FlagSet) <-chan error {
	errc := make(chan error)
	go func() {
		flags.VisitAll(func(flag *pflag.Flag) {
			for annot, val := range flag.Annotations {
				if annot == RequiredAnnotation && val[0] == "true" && flag.Value.String() == "" {
					errc <- promptForMissingFlag(flag)
				}
			}
		})
		close(errc)
	}()
	return errc
}

func promptForMissingFlag(flag *pflag.Flag) error {
	prompt := promptui.Prompt{
		Label:  fmt.Sprintf("Enter a value for %s", flag.Name),
		Stdout: os.Stderr,
	}

	val, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("missing required flag: %s", flag.Name)
	}

	if err := flag.Value.Set(val); err != nil {
		return fmt.Errorf("failed to set value for %s: %v", flag.Name, err)
	}
	return nil
}
