# Ephemeral IAM
A CLI tool that utilizes service account token generation to enabled users to
temporarily authenticate `gcloud` commands as a service account.  The intended
use-case for this tool is to restrict the permissions that users are granted
by default in their GCP organization while still allowing them to complete
management tasks that require escalated permissions.

> **Notice:** `ephemeral-iam` requires granting users the `Service Account Token Generator`
> role and does not include any controls to prevent users from using these
> privileges in contexts outside of `ephemeral-iam` in its current state.
> For more information on `ephemeral-iam`'s security considerations, refer to the
> [security considerations document](docs/security_considerations.md).

## Conceptual Overview
This section explains the basic process that happens when running the `eiam assume-privileges`
command.

Ephemeral IAM uses the `projects.serviceAccounts.generateAccessToken` method
to generate OAuth 2.0 tokens for service accounts which are then used in subsequent
API calls.  When a user runs the `assume-privileges` command, `eiam` makes a call
to generate an OAuth 2.0 token for the specified service account that expires
in 10 minutes. 

If the token was successfully generated, `eiam` then starts an
HTTPS proxy on the user's localhost. To enable the handling of HTTPS traffic,
a self-signed TLS certificate is generated for the proxy and stored for future
use.

Next, the active `gcloud` config is updated to forward all API calls through
the local proxy.

**Example updated configuration fields:**
```
[core]
  custom_ca_certs_file: [/path/to/eiam/config_dir/server.pem]
[proxy]
  address: [127.0.0.1]
  port: [8084]
  type: [http]
```

For the duration of the privileged session (either until the token expires or
when the user manually stops it with CTRL-C), all API calls made with `gcloud`
will be intercepted by the proxy which will replace the `Authorization` header
with the generated OAuth 2.0 token to authorize the request as the service account.

Once the session is over, `eiam` gracefully shuts down the proxy server and reverts
the users `gcloud` config to its original state.

## Installation
Instructions on how to install the `eiam` binary can be found in
[INSTALL.md](docs/INSTALL.md).

## Getting Started

### Generating Application Default Credentials
`ephemeral-iam` uses [Google Application Default Credentials](https://developers.google.com/identity/protocols/application-default-credentials)
for authorization credentials used in calling API endpoints and some commands
will fail if no ADCs are present. 

Generate Application Default Credentials:
```shell
$ gcloud auth application-default login
```

This will open a new window in your browser.  Login and select `Allow` when
prompted about the `Google Auth Library`.  If successful, the output in your
terminal should indicate where the file path that the ADC were written to.

### Help Commands
The root `eiam` invocation and each of its subcommands have their own help
commands. These commands can be used to gather more information about a command
and to explore the accepted arguments and flags.

Top-level `--help`
```
 $ eiam --help

╭────────────────────────────────────────────────────────────╮
│                                                            │
│                        Ephemeral IAM                       │
│  ──────────────────────────────────────────────────────    │
│  A CLI tool for temporarily escalating GCP IAM privileges  │
│  to perform high privilege tasks.                          │
│                                                            │
│      https://github.com/jessesomerville/ephemeral-iam      │
│                                                            │
╰────────────────────────────────────────────────────────────╯


╭────────────────────── Example usage ───────────────────────╮
│                                                            │
│                   Start privleged session                  │
│  ──────────────────────────────────────────────────────    │
│  $ eiam assumePrivileges \                                 │
│      -s example-svc@my-project.iam.gserviceaccount.com \   │
│      --reason "Emergency security patch (JIRA-1234)"       │
│                                                            │
│                                                            │
│                                                            │
│                     Run gcloud command                     │
│  ──────────────────────────────────────────────────────    │
│  $ eiam gcloud compute instances list --format=json \      │
│      -s example@my-project.iam.gserviceaccount.com \       │
│      -r "Reason"                                           │
│                                                            │
╰────────────────────────────────────────────────────────────╯

Usage:
  eiam [command]

Available Commands:
  assume-privileges     Configure gcloud to make API calls as the provided service account [alias: priv]
  config                Manage configuration values
  gcloud                Run a gcloud command with the permissions of the specified service account
  help                  Help about any command
  kubectl               Run a kubectl command with the permissions of the specified service account
  list-service-accounts List service accounts that can be impersonated [alias: list]
  query-permissions     Query current permissions on a GCP resource

Flags:
  -h, --help   help for eiam
  -y, --yes    Assume 'yes' to all prompts

Use "eiam [command] --help" for more information about a command.
```

Subcommand `--help`
```
 $ eiam assume-privileges --help

The "assume-privileges" command fetches short-lived credentials for the provided service Account
and configures gcloud to proxy its traffic through an auth proxy. This auth proxy sets the
authorization header to the OAuth2 token generated for the provided service account. Once
the credentials have expired, the auth proxy is shut down and the gcloud config is restored.

The reason flag is used to add additional metadata to audit logs.  The provided reason will
be in 'protoPatload.requestMetadata.requestAttributes.reason'.

Usage:
  eiam assume-privileges [flags]

Aliases:
  assume-privileges, priv

Examples:

eiam assume-privileges \
  --serviceAccountEmail example@my-project.iam.gserviceaccount.com \
  --reason "Emergency security patch (JIRA-1234)"

Flags:
  -h, --help                         help for assume-privileges
  -p, --project string               The GCP project. Inherits from the active gcloud config by default (default "rigup-sandbox")
  -R, --reason string                A detailed rationale for assuming higher permissions
  -s, --serviceAccountEmail string   The email address for the service account

Global Flags:
  -y, --yes   Assume 'yes' to all prompts
```

### Tutorial
To better familiarize yourself with `ephemeral-iam` and how it works, you can
follow [the tutorial provided in the documentation](docs/tutorial).


# TODO
- [ ] Add global common flags (e.g. project)
- [ ] Add NOTICE file
- [ ] Write `tutorial.md`
- [ ] Unit tests
- [ ] Build and publish release binaries
- [ ] Explore the possiblility of using a dedicated gcloud config to prevent potential issues around modifying the users config