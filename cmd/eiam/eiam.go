package main

import (
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/cmd"
	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

func main() {
	cmd := cmd.NewEphemeralIamCommand()
	util.CheckError(cmd.Execute())
}
