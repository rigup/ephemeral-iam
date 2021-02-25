package main

import (
	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/cmd"
	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

func main() {
	util.CheckError(util.CheckDependencies())

	cmd := cmd.NewEphemeralIamCommand()
	util.CheckError(cmd.Execute())
}
