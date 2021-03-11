# Installation

### Option 1: Install a release binary

1. Download the binary for your OS from the [releases page](https://github.com/jessesomerville/ephemeral-iam/releases)

2. OPTIONAL: Download the `checksums.txt` and `checksums.txt.sig` files to verify the integrity of the archive

```shell
# Check the signature of the checksums
$ gpg --verify checksums.txt.sig

# Check the checksum of the downloaded archive
$ shasum -a 256 ephemeral-iam_${VERSION}_${ARCH}.tar.gz
ed395b9acb603ad87819ab05b262b9d725186d9639c09dd2545898ed308720f9  ephemeral-iam_${VERSION}_${ARCH}.tar.gz

$ cat checksums.txt | grep ephemeral-iam_${VERSION}_${ARCH}.tar.gz
ed395b9acb603ad87819ab05b262b9d725186d9639c09dd2545898ed308720f9  ephemeral-iam_${VERSION}_${ARCH}.tar.gz
```

3. Extract the downloaded archive

```
$ tar -xvf ephemeral-iam_${VERSION}_${ARCH}.tar.gz
```

4. Move the `eiam` binary into your path:

```shell
$ mv ./eiam /usr/local/bin/
```

5. Verify the installation

```shell
$ eiam version
INFO    ephemeral-iam 
```

> **NOTE:** If you are on macOS and you get an error due to the binary being from an unknown publisher, remove
> the quarantine label from the binary:

```shell
$ xattr -d com.apple.quarantine /path/to/eiam
```

### Option 2: Installing with `go get`
#### Prerequisites

1. You must have [`go`](https://golang.org/doc/install) installed.

2. `ephemeral-iam` has only been tested on **macOS Big Sur 11.2** and **Ubuntu 20.04**.
   It is likely to work on other versions of macOS and Linux.  If you discover
   a bug while using this tool on on other platforms/architecture, please open a
   new issue.

#### Configure Go Environment

You will want to configure your `GOPATH` environment variable and add it to
your `PATH`.  Additionally, you may want to add them to your shell configuration
for convenience (`.bashrc`, `.zshrc`, etc). More information about the `GOPATH`
variable can be found [here](https://github.com/golang/go/wiki/GOPATH).

```shell
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin
```

#### Install the package
```shell
# Install using go get
$ GO111MODULE="on" go get github.com/jessesomerville/ephemeral-iam/...
```

This will create the `eiam` binary in your `$GOPATH/bin` directory. You can
confirm this was successful by running `which eiam` or invoking the `help`
command as shown below.

```shell
# Ensure the eiam binary was successfully added to your PATH
$ eiam --help
```