# Basic Commands
The following are basic commands offered by `eiam`

## Get Version

```
$ eiam version
INFO    ephemeral-iam vX.Y.Z
```

## ephemeral-iam Configuration
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
  certfile: /Users/example/Library/Application Support/ephemeral-iam/server.pem
  keyfile: /Users/example/Library/Application Support/ephemeral-iam/server.key
  logdir: /Users/example/Library/Application Support/ephemeral-iam/log
  proxyaddress: 127.0.0.1
  proxyport: "8084"
  verbose: true
binarypaths:
  gcloud: /Users/example/google-cloud-sdk/bin/gcloud
  kubectl: /usr/local/bin/kubectl
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

┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ Key                            ┃ Description                                 ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┩
│ authproxy.certfile             │ The path to the auth proxy's TLS            │
│                                │ certificate                                 │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ authproxy.keyfile              │ The path to the auth proxy's x509 key       │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ authproxy.logdir               │ The directory that auth proxy logs will be  │
│                                │ written to                                  │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ authproxy.proxyaddress         │ The address that the auth proxy is hosted   │
│                                │ on                                          │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ authproxy.proxyport            │ The port that the auth proxy runs on        │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ authproxy.verbose              │ When set to 'true', verbose output for      │
│                                │ proxy logs will be enabled                  │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ binarypaths.gcloud             │ The path to the gcloud binary on your       │
│                                │ filesystem                                  │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ binarypaths.kubectl            │ The path to the kubectl binary on your      │
│                                │ filesystem                                  │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ logging.format                 │ The format for which to write console logs  │
│                                │ Can be either 'json', 'text', or 'debug'    │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ logging.level                  │ The logging level to write to the console   │
│                                │ Can be one of 'trace', 'debug', 'info',     │
│                                │ 'warn', 'error', 'fatal', or 'panic'        │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ logging.disableleveltruncation │ When set to 'true', the level indicator for │
│                                │ logs will not be trucated                   │
├────────────────────────────────┼─────────────────────────────────────────────┤
│ logging.padleveltext           │ When set to 'true', output logs will align  │
│                                │ evenly with their output level indicator    │
└────────────────────────────────┴─────────────────────────────────────────────┘
```

### Set a configuration value

```
$ eiam config set logging.level debug
INFO    Updated logging.level from info to debug
```