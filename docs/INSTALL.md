# Installation

## Option 1: Install a release binary

TODO

## Option 2: Installing with `go get`

### Prerequisites

1. You must have [`go`](https://golang.org/doc/install) installed.

2. `ephemeral-iam` has only been tested on **macOS Big Sur 11.2** and **Ubuntu 20.04**.
   It is likely to work on other versions of macOS and Linux.  If you discover
   a bug while using this tool on on other platforms/architecture, please open a
   new issue.

### Configure Go Environment

You will want to configure your `GOPATH` environment variable and add it to
your `PATH`.  Additionally, you may want to add them to your shell configuration
for convenience (`.bashrc`, `.zshrc`, etc). More information about the `GOPATH`
variable can be found [here](https://github.com/golang/go/wiki/GOPATH).

```shell
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin
```

### Install the package
```shell
# Install using go get
$ go get github.com/jessesomerville/ephemeral-iam/...
```

This will create the `eiam` binary in your `$GOPATH/bin` directory. You can
confirm this was successful by running `which eiam` or invoking the `help`
command as shown below.

```shell
# Ensure the eiam binary was successfully added to your PATH
$ eiam --help
```