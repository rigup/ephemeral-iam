package eiamplugin

import (
	"fmt"

	"github.com/spf13/cobra"
)

type EphemeralIamPlugin struct {
	*cobra.Command
	name    string
	desc    string
	version string
}

func (eiamp *EphemeralIamPlugin) Name() string {
	return eiamp.name
}

func (eiamp *EphemeralIamPlugin) Desc() string {
	return eiamp.desc
}

func (eiamp *EphemeralIamPlugin) Version() string {
	return eiamp.version
}

func New() *EphemeralIamPlugin {
	return &EphemeralIamPlugin{
		Command: newCommand(),
		name:    "plugin-template",
		desc:    "Baseline template for new plugins for ephemeral-iam",
		version: "v0.0.0",
	}
}

func newCommand() *cobra.Command {
	return &cobra.Command{
		Use: "plugin-template",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TODO: Tutorial on writing plugins")
		},
	}
}
