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

package eiamplugin

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
	eiamplugin "github.com/rigup/ephemeral-iam/pkg/plugins"
)

type RootCommand struct {
	Plugins []*eiamplugin.EphemeralIamPlugin
	cobra.Command
}

func (rc *RootCommand) loadPlugin(pluginPath string) (*eiamplugin.EphemeralIamPlugin, bool, error) {
	pluginLib, err := plugin.Open(pluginPath)
	if err != nil {
		if serr := errorsutil.CheckPluginError(err); serr != nil {
			return nil, false, serr
		}
		return nil, false, nil
	}

	newPlugin, err := pluginLib.Lookup("Plugin")
	if err != nil {
		return nil, false, errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: fmt.Sprintf("The plugin %s is missing the EphemeralIamPlugin symbol", pluginPath),
			Err: err,
		}
	}
	if p, ok := newPlugin.(**eiamplugin.EphemeralIamPlugin); ok {
		return *p, true, nil
	}
	err = fmt.Errorf("plugin of type %T should be *eiamplugin.EphemeralIamPlugin", newPlugin)
	return nil, false, errorsutil.EiamError{
		Log: util.Logger.WithError(err),
		Msg: fmt.Sprintf("Failed to load plugin %s", pluginPath),
		Err: err,
	}
}

func (rc *RootCommand) LoadPlugins() error {
	configDir := appconfig.GetConfigDir()

	pluginPaths := []string{}
	err := filepath.Walk(filepath.Join(configDir, "plugins"), func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".so") {
			pluginPaths = append(pluginPaths, path)
		}
		return nil
	})
	if err != nil {
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: "Failed to list files in plugins directory",
			Err: err,
		}
	}

	loadedPlugins := []*eiamplugin.EphemeralIamPlugin{}
	for _, path := range pluginPaths {
		if p, loaded, err := rc.loadPlugin(path); err != nil {
			return err
		} else if loaded {
			rc.AddCommand(p.Command)
			p.Path = path
			loadedPlugins = append(loadedPlugins, p)
		}
	}
	if len(rc.Plugins) != 0 {
		util.Logger.Debugf("Successfully loaded %d plugins", len(pluginPaths))
		rc.Plugins = loadedPlugins
	}
	return nil
}

func (rc *RootCommand) PrintPlugins() {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 4, 4, ' ', 0)
	fmt.Fprintln(w, "\nPLUGIN\tVERSION\tDESCRIPTION")
	for _, p := range rc.Plugins {
		fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Version, p.Desc)
	}
	fmt.Fprintln(w)
	w.Flush()

	fmt.Println(buf.String())
}
