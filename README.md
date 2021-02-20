# ephemeral-iam
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

> **TODO:** I want to give an overview of how this all works under the hood.
> Especially around the gcloud configuration changes

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
Utility for granting short-lived, privileged access to GCP APIs.

Usage:
  gcp-iam-escalate [command]

Available Commands:
  assumePrivileges    Configure gcloud to make API calls as the provided service account
  editConfig          Edit configuration values
  gcloud              Run a gcloud command with the permissions of the specified service account
  help                Help about any command
  kubectl             Run a kubectl command with the permissions of the specified service account
  listServiceAccounts List service accounts that can be impersonated

Flags:
  -h, --help   help for gcp-iam-escalate

Use "gcp-iam-escalate [command] --help" for more information about a command.
```

Subcommand `--help`
```
 $ ./eiam assumePrivileges --help

The "assumePrivileges" command fetches short-lived credentials for the provided service Account
and configures gcloud to proxy its traffic through an auth proxy. This auth proxy sets the
authorization header to the OAuth2 token generated for the provided service account. Once
the credentials have expired, the auth proxy is shut down and the gcloud config is restored.

The reason flag is used to add additional metadata to audit logs.  The provided reason will
be in 'protoPayload.requestMetadata.requestAttributes.reason'.

Example:
  gcp_iam_escalate assumePrivileges \
      --serviceAccountEmail example@my-project.iam.gserviceaccount.com \
      --reason "Emergency security patch (JIRA-1234)"

Usage:
  gcp-iam-escalate assumePrivileges [flags]

Flags:
  -h, --help                         help for assumePrivileges
      --reason string                A detailed rationale for assuming higher permissions (required)
      --serviceAccountEmail string   The email address for the service account to impersonate (required)
```

### Tutorial
To better familiarize yourself with `ephemeral-iam` and how it works, you can
follow [the tutorial provided in the documentation](docs/tutorial.md).


# TODO
- [ ] Finish documentation (`Concepts`, `security_considerations.md`, and `tutorial.md`)
- [ ] Unit tests
- [ ] Build and publish release binaries
- [ ] Explore the possiblility of using a dedicated gcloud config to prevent potential issues around modifying the users config