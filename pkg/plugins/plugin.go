package eiamplugin

import "github.com/spf13/cobra"

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
