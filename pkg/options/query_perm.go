package options

import (
	"github.com/spf13/pflag"

	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
)

// Flag names and shorthands
var (
	ComputeInstanceFlag = flagName{"instance", "i"}
	PubSubTopicFlag     = flagName{"topic", "t"}
	StorageBucketFlag   = flagName{"bucket", "b"}
)

// AddComputeInstanceFlag adds the --instance/-i flag to the command
func AddComputeInstanceFlag(fs *pflag.FlagSet, instance *string, required bool) {
	fs.StringVarP(instance, ComputeInstanceFlag.Name, ComputeInstanceFlag.Shorthand, "", "The name of the compute instance")
	if required {
		if err := fs.SetAnnotation(ComputeInstanceFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}

// AddPubSubTopicFlag adds the --topic/-t flag to the command
func AddPubSubTopicFlag(fs *pflag.FlagSet, topic *string, required bool) {
	fs.StringVarP(topic, PubSubTopicFlag.Name, PubSubTopicFlag.Shorthand, "", "The name of the Pub/Sub topic")
	if required {
		if err := fs.SetAnnotation(PubSubTopicFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}

// AddStorageBucketFlag adds the --bucket/-b flag to the command
func AddStorageBucketFlag(fs *pflag.FlagSet, bucket *string, required bool) {
	fs.StringVarP(bucket, StorageBucketFlag.Name, StorageBucketFlag.Shorthand, "", "The name of the storage bucket")
	if required {
		if err := fs.SetAnnotation(StorageBucketFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}
