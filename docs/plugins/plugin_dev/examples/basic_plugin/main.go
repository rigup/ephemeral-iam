// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"math/rand"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
)

const (
	name    = "basic-plugin"
	desc    = "An example of a basic eiam plugin command"
	version = "v0.0.1"
)

// BasicPlugin is the implementation of the ephemeral-iam plugin interface.
type BasicPlugin struct {
	// Logger is the logger for the plugin to use to send log entries to eiam
	// to be formatted and output to the user.
	Logger hclog.Logger
}

// GetInfo is the function that eiam invokes to get metadata about the plugin.
func (p *BasicPlugin) GetInfo() (n, d, v string, err error) {
	return name, desc, version, nil
}

// Run is the function that eiam uses to invoke the plugin command.
func (p *BasicPlugin) Run() error {
	rootCmd := newRootCmd(p)
	return rootCmd.Execute()
}

func newRootCmd(p *BasicPlugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: desc,
		// Plugins should use the RunE/PreRunE fields and return their errors
		// to be handled by eiam.
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Logger.Info("This is printed in the same format as other `eiam` INFO logs")
			p.Logger.Error("This is an error message")
			rand.Seed(time.Now().UnixNano())
			if rand.Intn(2) == 1 {
				return errors.New("this is an example error returned to eiam")
			}
			return nil
		},
	}
	return cmd
}
