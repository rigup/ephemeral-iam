package eiamplugin

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
)

type EphemeralIamPlugin struct {
	*cobra.Command
	Name    string
	Desc    string
	Version string
}

func Logger() *logrus.Logger {
	return util.Logger
}
