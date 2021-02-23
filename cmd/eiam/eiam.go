package main

import (
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/cmd"
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

func main() {
	cmd := cmd.NewEphemeralIamCommand()
	eiamutil.CheckError(cmd.Execute())
}
