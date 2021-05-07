# ephemeral-iam Plugins
Custom functionality can be added to `ephemeral-iam` by the way of plugins.
Plugins are [cobra](https://github.com/spf13/cobra) Commands that communicate
with `ephemeral-iam` over gRPC using Hashicorp's `go-plugin` package.

## Installing a plugin
Plugins are loaded from the `/path/to/config/ephemeral-iam/plugins` directory.
To install a plugin, you just place the plugin's binary in that directory and
`eiam` will automatically discover and load it.

If the plugin you want to install is hosted in a Github repo and the binary is
published as a release in the repository, you can install the plugin using the
`eiam plugin install` command:

```
$ eiam plugin install --url github.com/user/repo-name
```

## Developing a new plugin
Details on how to develop a custom plugin can be found in the [plugin-dev](plugin-dev)
directory.