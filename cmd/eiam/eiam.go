package main

import (
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/cmd"
	errorsutil "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/errors"
)

func main() {
	cmd := cmd.NewEphemeralIamCommand()
	errorsutil.CheckError(cmd.Execute())
}
