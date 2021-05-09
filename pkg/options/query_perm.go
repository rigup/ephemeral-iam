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
	"github.com/spf13/pflag"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
)

// Flag names and shorthands.
var (
	// ComputeInstanceFlag sets the compute instance to use for a command.
	ComputeInstanceFlag = flagName{"instance", "i"}

	// PubSubTopicFlag sets the Pub/Sub topic to use for a command.
	PubSubTopicFlag = flagName{"topic", "t"}

	// StorageBucketFlag sets the GCS bucket to use for a command.
	StorageBucketFlag = flagName{"bucket", "b"}
)

// AddComputeInstanceFlag adds the --instance/-i flag to the command.
func AddComputeInstanceFlag(fs *pflag.FlagSet, instance *string, required bool) {
	fs.StringVarP(
		instance,
		ComputeInstanceFlag.Name,
		ComputeInstanceFlag.Shorthand,
		"",
		"The name of the compute instance",
	)
	if required {
		if err := fs.SetAnnotation(ComputeInstanceFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}

// AddPubSubTopicFlag adds the --topic/-t flag to the command.
func AddPubSubTopicFlag(fs *pflag.FlagSet, topic *string, required bool) {
	fs.StringVarP(topic, PubSubTopicFlag.Name, PubSubTopicFlag.Shorthand, "", "The name of the Pub/Sub topic")
	if required {
		if err := fs.SetAnnotation(PubSubTopicFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}

// AddStorageBucketFlag adds the --bucket/-b flag to the command.
func AddStorageBucketFlag(fs *pflag.FlagSet, bucket *string, required bool) {
	fs.StringVarP(
		bucket,
		StorageBucketFlag.Name,
		StorageBucketFlag.Shorthand,
		"",
		"The name of the storage bucket",
	)
	if required {
		if err := fs.SetAnnotation(StorageBucketFlag.Name, RequiredAnnotation, []string{"true"}); err != nil {
			util.Logger.Fatalf("failed to set required annotation on flag: %v", err)
		}
	}
}
