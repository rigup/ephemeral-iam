package options

import (
	"github.com/spf13/pflag"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/gcpclient"
)

// Flag annotation strings
const (
	RequiredAnnotation = "eiam_required_flag"
)

// YesOption designates whether to prompt for confirmation or not
var YesOption = false

// Flag names and shorthands
var (
	ProjectFlag             = flagName{"project", "p"}
	ReasonFlag              = flagName{"reason", "R"}
	RegionFlag              = flagName{"region", "r"}
	ServiceAccountEmailFlag = flagName{"serviceAccountEmail", "s"}
	YesFlag                 = flagName{"yes", "y"}
	ZoneFlag                = flagName{"zone", "z"}
)

type flagName struct {
	Name      string
	Shorthand string
}

// CmdConfig holds the values passed to a command
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

// AddPersistentFlags add persistent flags to the root command
func AddPersistentFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&YesOption, YesFlag.Name, YesFlag.Shorthand, YesOption, "Assume 'yes' to all prompts")
}

// AddProjectFlag adds the --project/-p flag to the command
func AddProjectFlag(fs *pflag.FlagSet, project *string) {
	defaultVal, err := gcpclient.GetCurrentProject()
	util.CheckError(err)

	fs.StringVarP(project, ProjectFlag.Name, ProjectFlag.Shorthand, defaultVal, "The GCP project. Inherits from the active gcloud config by default")
}

// AddRegionFlag adds the --region/-r flag to the command
func AddRegionFlag(fs *pflag.FlagSet, region *string, required bool) {
	defaultVal, err := gcpclient.GetCurrentRegion()
	util.CheckError(err)

	fs.StringVarP(region, RegionFlag.Name, RegionFlag.Shorthand, defaultVal, "The GCP region. Inherits from the active gcloud config by default")
	if required {
		fs.SetAnnotation(RegionFlag.Name, RequiredAnnotation, []string{"true"})
	}
}

// AddZoneFlag adds the --zone/-z flag to the command
func AddZoneFlag(fs *pflag.FlagSet, zone *string, required bool) {
	defaultVal, err := gcpclient.GetCurrentZone()
	util.CheckError(err)

	fs.StringVarP(zone, ZoneFlag.Name, ZoneFlag.Shorthand, defaultVal, "The GCP zone. Inherits from the active gcloud config by default")
	if required {
		fs.SetAnnotation(ZoneFlag.Name, RequiredAnnotation, []string{"true"})
	}
}

// AddServiceAccountEmailFlag adds the --serviceAccountEmail/-s flag
func AddServiceAccountEmailFlag(fs *pflag.FlagSet, serviceAccountEmail *string, required bool) {
	fs.StringVarP(serviceAccountEmail, ServiceAccountEmailFlag.Name, ServiceAccountEmailFlag.Shorthand, "", "The email address for the service account")
	if required {
		fs.SetAnnotation(ServiceAccountEmailFlag.Name, RequiredAnnotation, []string{"true"})
	}
}

// AddReasonFlag adds the --reason/-R flag
func AddReasonFlag(fs *pflag.FlagSet, reason *string, required bool) {
	fs.StringVarP(reason, ReasonFlag.Name, ReasonFlag.Shorthand, "", "A detailed rationale for assuming higher permissions")
	if required {
		fs.SetAnnotation(ReasonFlag.Name, RequiredAnnotation, []string{"true"})
	}
}

// CheckRequired ensures that a command's required flags have been set
func CheckRequired(flag *pflag.Flag) {
	for annot, val := range flag.Annotations {
		if annot == RequiredAnnotation && val[0] == "true" {
			if flag.Value.String() == "" {
				util.Logger.Fatalf("Missing required flag: %s", flag.Name)
			}
		}
	}
}
