package main

import (
	"github.com/jessesomerville/ephemeral-iam/cmd"
	errorsutil "github.com/jessesomerville/ephemeral-iam/internal/errors"
)

func main() {
	cmd, err := cmd.NewEphemeralIamCommand()
	errorsutil.CheckError(err)
	errorsutil.CheckError(cmd.Execute())
}
