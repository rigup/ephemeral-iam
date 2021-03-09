# Basic Commands
The following are basic commands offered by `eiam`

## Get Version

```
$ eiam version
INFO    ephemeral-iam v0.0.2
```

## Ephemeral IAM Configuration
When you run `eiam` for the first time it will generate a baseline configuration that works for most use cases.
There are a few commands that you can use to view and edit this config based on your needs.

```
$ eiam config --help
Manage configuration values

Usage:
  eiam config [command]

Available Commands:
  info        Print information about config fields
  print       Print the current configuration
  set         Set the value of a provided config item
  view        View the value of a provided config item

Flags:
  -h, --help   help for config

Global Flags:
  -y, --yes   Assume 'yes' to all prompts

Use "eiam config [command] --help" for more information about a command.
```

### Print the current configuration

```
$ eiam config print

authproxy:
  logdir: /Users/jsomerville/Library/Application Support/ephemeral-iam/log
  proxyaddress: 127.0.0.1
  proxyport: "8084"
  verbose: false
  writetofile: false
logging:
  disableleveltruncation: true
  format: text
  level: info
  padleveltext: true
```

### View a single configuration item

```
$ eiam config view authproxy.proxyaddress
INFO    authproxy.proxyaddress: 127.0.0.1
```

### Get information about configuration fields

```
$ eiam config info

┏━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ Key                    ┃ Description                                         ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┩
│ AuthProxy.ProxyAddress │ The address to the auth proxy. You shouldn't need   │
│                        │ to update this                                      │
├────────────────────────┼─────────────────────────────────────────────────────┤
│ AuthProxy.ProxyPort    │ The port to run the auth proxy on                   │
├────────────────────────┼─────────────────────────────────────────────────────┤
│ AuthProxy.Verbose      │ Enables verbose logging output from the auth proxy  │
├────────────────────────┼─────────────────────────────────────────────────────┤
│ AuthProxy.WriteToFile  │ Enables writing auth proxy logs to a log file       │
├────────────────────────┼─────────────────────────────────────────────────────┤
│ AuthProxy.LogDir       │ The directory to write auth proxy logs to           │
├────────────────────────┼─────────────────────────────────────────────────────┤
│ Logging.Format         │ The format for the console logs.                    │
│                        │ Can be either 'json' or 'text'                      │
├────────────────────────┼─────────────────────────────────────────────────────┤
│ Logging.Level          │ The logging level to write to the console.          │
│                        │ Can be one of "trace", "debug", "info", "warn",     │
│                        │ "error", "fatal", "panic"                           │
└────────────────────────┴─────────────────────────────────────────────────────┘
```

### Set a configuration value

```
$ eiam config set logging.level debug
INFO    Updated logging.level from info to debug
```