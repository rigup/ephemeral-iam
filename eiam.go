package main

import (
	"github.com/rigup/ephemeral-iam/cmd"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

func main() {
	cmd, err := cmd.NewEphemeralIamCommand()
	errorsutil.CheckError(err)
	errorsutil.CheckError(cmd.Execute())
}
