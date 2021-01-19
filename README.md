# ephemeral-iam
A CLI tool for temporarily escalating GCP IAM privileges to perform high privilege tasks.


# Installation

### Minimum Requirements
 - Linux or macOS
 - Go 1.15 or higher

### Suggested
- Add your _\$GOPATH/bin_ into your _\$PATH_ ([instructions](https://github.com/golang/go/wiki/GOPATH))

Install the package
```
go get github.com/jessesomerville/ephemeral-iam
```

# Usage

```
$ ephemeral-iam help

	A proof-of-concept CLI tool that demonstrates gaining short-term access to GCP APIs using
	short-lived service account credentials.

Usage:
  ephemeral-iam [command]

Available Commands:
  assumePrivileges    Configure gcloud to make API calls as the provided service account
  editConfig          Edit configuration values
  help                Help about any command
  listServiceAccounts List service accounts that can be impersonated

Flags:
  -h, --help   help for ephemeral-iam

Use "ephemeral-iam [command] --help" for more information about a command.
```

# TODO
- [ ] Generate certs with golang instead of bash script
- [ ] Add integration with Slack for alerting.  This could also be implemented in GCP using audit log export filters and cloud functions
- [ ] Write tests