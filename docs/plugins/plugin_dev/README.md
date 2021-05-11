# Plugin Development
`ephemeral-iam` plugins utilize Hashicorp's [go-plugin](https://github.com/hashicorp/go-plugin)
package and [spf13's cobra package](https://github.com/spf13/cobra). See the [examples](examples)
directory for examples of valid of `ephemeral-iam` plugins.

 - [Basic plugin](examples/basic_plugin)
 - [Command with flags](examples/command_flags)
 - [Plugin with subcommands](examples/subcommands)

## Plugin Interface
Plugins are expected to implement the `EIAMPlugin` interface.  This interface
includes two functions: `GetInfo` and `Run`.

`GetInfo` is the function that `ephemeral-iam` uses to get metadata about a plugin.
The function returns the plugin's name, description, and version.

```go
func (p *EIAMPlugin) GetInfo() (name, desc, version string, err error) {
    return "example", "This is an example", "v0.0.1", nil
} 
```

`Run` is the function that `ephemeral-iam` invokes when a user uses the plugin's
command.  Any errors returned will be propagated back to eiam to handle.

```go
func (p *EIAMPlugin) Run() error {
	cmd := &cobra.Command{
		Use:   name,
		Short: desc,
		// Plugins should use the RunE/PreRunE fields and return their errors
		// to be handled by eiam.
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Logger.Info("This is printed in the same format as other `eiam` INFO logs")
			p.Logger.Error("This is an error message")
			rand.Seed(time.Now().UnixNano())
			if rand.Intn(2) == 1 {
				return errors.New("this is an example error returned to eiam")
			}
			return nil
		},
	}
    return cmd.Execute()
}
```