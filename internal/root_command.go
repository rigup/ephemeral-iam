package eiam_plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
	eiamplugin "github.com/jessesomerville/ephemeral-iam/pkg/plugins"
)

type RootCommand struct {
	cobra.Command
}

func (rc *RootCommand) loadPlugin(pluginPath string) (*eiamplugin.EphemeralIamPlugin, error) {
	pluginLib, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	newPlugin, err := pluginLib.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("%s is missing the EphemeralIamPlugin symbol", pluginPath)
	}
	pluginDef := newPlugin.(func() interface{})()
	if p, ok := pluginDef.(*eiamplugin.EphemeralIamPlugin); ok {
		return p, nil
	}
	return nil, fmt.Errorf("ersdrs")
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

	for _, path := range pluginPaths {
		if p, err := rc.loadPlugin(path); err != nil {
			return err
		} else {
			fmt.Printf("Plugin Name: %s\nPlugin Desc: %s\nVersion: %s\n\n", p.Name(), p.Desc(), p.Version())
		}
	}
	if len(pluginPaths) != 0 {
		util.Logger.Infof("Successfully loaded %d plugins", len(pluginPaths))
	}
	return nil
}
