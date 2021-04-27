# ephemeral-iam Plugins
Plugins for ephemeral-iam allow you to extend `eiam`'s functionality in the form of commands.
Plugins are `.so` files (Golang dynamic libraries) and stored in the `plugins` directory
of your `eiam` configuration folder.

## Installing a plugin
To install a plugin, take the `.so` file and place it in the `plugins` directory of your
`eiam` configuration folder.  `eiam` will automatically load any valid plugins in that
directory and the commands added by those plugins will be listed when you run:
`eiam --help`.

## Developing a new plugin
For information on how to develop an ephemeral-iam plugin, click [here](plugin_dev).