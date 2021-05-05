# Plugin Development
ephemeral-iam plugins utilize Hashicorp's [go-plugin](https://github.com/hashicorp/go-plugin)
package and [spf13's cobra package](https://github.com/spf13/cobra). See the
[examples](examples) directory for more details on specific aspects of
`ephemeral-iam` plugins.

 - [Basic plugin](examples/basic_plugin)
 - [Command with flags](examples/command_flags)
 - [Plugin with subcommands](examples/subcommands)