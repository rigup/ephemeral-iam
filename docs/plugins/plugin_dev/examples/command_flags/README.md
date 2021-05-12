# Plugin Command Flags
Plugins can use the flags that are used by native `ephemeral-iam` commands. To
see an example of this, reference the [Command with flags](examples/command_flags)
example.

Here are the available flags and their intended usage.

| Name                    | CLI Format                     | Description                                                                                                               |
|-------------------------|--------------------------------|---------------------------------------------------------------------------------------------------------------------------|
| ComputeInstanceFlag     | `--instance`/`-i`              | The name of a compute instance                                                                                            |
| ProjectFlag             | `--project`/`-p`               | The GCP project. Inherits from the active gcloud config by default                                                        |
| PubSubTopicFlag         | `--topic`/`-t`                 | The name of a Pub/Sub topic                                                                                               |
| ReasonFlag              | `--reason`/`-R`                | The reason for running a command. `ephemeral-iam` uses this with the `WithRequestReason` option when creating API clients |
| RegionFlag              | `--region`/`-r`                | The GCP region. Inherits from the active gcloud config by default                                                         |
| ServiceAccountEmailFlag | `--service-account-email`/`-s` | The email address of a service account                                                                                    |
| StorageBucketFlag       | `--bucket`/`-b`                | The name of a Storage Bucket                                                                                              |
| YesFlag                 | `--yes`/`-y`                   | Assume 'yes' to all prompts                                                                                               |
| ZoneFlag                | `--zone`/`-z`                  | The GCP zone. Inherits from the active gcloud config by default                                                           |

Any of these flags can be marked as required by setting the last parameter in the
function call to add them to a command to true.  To then check for any missing
required flags, call `CheckRequired` in the command's `PreRunE` function.

**Example:**
```go

import (
    "github.com/rigup/ephemeral-iam/pkg/options"
    "github.com/spf13/cobra"
)

...

func pluginFuncWithEiamFlags(p *MyPlugin) *cobra.Command {
    var (
        instance string
        bucket   string
    )

    cmd := &cobra.Command{
        Use: "example",
        PreRunE: func(cmd *cobra.Command, args []string) error {
            // Check that the compute instance flag was provided
            return options.CheckRequired(cmd.Flags())
        },
        RunE: func(cmd *cobra.command, args []string) error {
            p.Log.Info("You provided the requied instance flag", "instance", instance)
            if bucket != "" {
                p.Log.Info("You provided the optional bucket flag", "bucket", bucket)
            }
            return nil
        }
    }
    // Add the `--instance`/`-i` flag and make it required
    options.AddComputeInstanceFlag(cmd.Flags(), &instance, true)
    // Add the `--bucket`/`-b` flag and make it optional
    options.AddStorageBucketFlag(cmd.Flags(), &bucket, false)

	return cmd
}
```

## Adding custom flags to plugin commands
Custom flags can be added to plugin commands just like any other Cobra command
as long as the name/shortform does not conflict with an existing flag.

**Example:**
```go
func pluginFuncWithEiamFlags() *cobra.Command {
    cmd := &cobra.Command{
        Use: "example",
        RunE: func(cmd *cobra.command, args []string) error {
            if Verbose {
                fmt.Println("Verbose output enabled")
            }
            return nil
        }
    }

    cmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}
```