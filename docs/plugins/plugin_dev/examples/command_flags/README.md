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

## Adding custom flags to plugin commands
Custom flags can be added to plugin commands just like any other Cobra command
as long as the name/shortform does not conflict with an existing flag.

**Example:**
```go
var (
	Plugin = &eiamplugin.EphemeralIamPlugin{
		// Command defines the top-level command that will be added to eiam.
		// It is an instance of cobra.Command (https://pkg.go.dev/github.com/spf13/cobra#Command)
		Command: pluginFuncWithEiamFlags(),
		Name:    "Plugin with command flags",
		Desc:    "This is an example plugin with command flags",
		Version: "v0.0.1",
	}

    Verbose bool
)

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