package eiam_plugin

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
	eiamplugin "github.com/jessesomerville/ephemeral-iam/pkg/plugins"
)

type RootCommand struct {
	cobra.Command
	Plugins []*eiamplugin.EphemeralIamPlugin
}

// TODO: custom error type for plugins
func (rc *RootCommand) loadPlugin(pluginPath string) (*eiamplugin.EphemeralIamPlugin, error) {
	pluginLib, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	newPlugin, err := pluginLib.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("%s is missing the EphemeralIamPlugin symbol", pluginPath)
	}
	if p, ok := newPlugin.(**eiamplugin.EphemeralIamPlugin); ok {
		return *p, nil
	}
	return nil, fmt.Errorf("failed to load plugin %s: plugin of type %T should be *eiamplugin.EphemeralIamPlugin", pluginPath, newPlugin)
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
		return err
	}

	loadedPlugins := []*eiamplugin.EphemeralIamPlugin{}
	for _, path := range pluginPaths {
		if p, err := rc.loadPlugin(path); err != nil {
			return err
		} else {
			rc.AddCommand(p.Command)
			p.Path = path
			loadedPlugins = append(loadedPlugins, p)
		}
	}
	if len(pluginPaths) != 0 {
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
