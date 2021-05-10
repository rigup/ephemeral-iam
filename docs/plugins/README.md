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
`eiam plugins install` command:

```
$ eiam plugins install --url github.com/user/repo-name
```

### Plugin stored in a private repository
If the plugin is hosted in a private repository, you need to provide `ephemeral-iam`
with a Github personal access token to authenticate with. You can use the 
`eiam plugins auth` commands to add, list, and remove these access tokens.

**Add tokens:**
```
$ eiam plugins auth add --name "personal-token"
INFO    Adding token with the name personal-token
✔ Enter your Github Personal Access Token: : ●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●

$ eiam plugins auth add --name "organization-token"
INFO    Adding token with the name organization-token
✔ Enter your Github Personal Access Token: : ●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●
```

**List tokens:**
```
$ eiam plugins auth list

 GITHUB ACCESS TOKENS
----------------------
personal-token
organization-token
```

**Install with authentication:**
```
$ eiam plugins install --url github.com/user/repo-name --token personal-token
```

## Developing a new plugin
Details on how to develop a custom plugin can be found in the [plugin-dev](plugin-dev)
directory.