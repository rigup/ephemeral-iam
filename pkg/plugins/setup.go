package eiamplugin

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
)

var PluginDir string

func init() {
	configDir := appconfig.GetConfigDir()
	PluginDir = filepath.Join(configDir, "plugins")

	if err := createPluginDir(); err != nil {
		log.Fatalf("Setup error: %v", err)
	}
}

func createPluginDir() error {
	if _, err := os.Stat(PluginDir); os.IsNotExist(err) {
		util.Logger.Debugf("Creating plugin directory: %s", PluginDir)
		if err := os.MkdirAll(PluginDir, 0o755); err != nil {
			return fmt.Errorf("failed to create plugin directory %s: %v", PluginDir, err)
		}
	} else if err != nil {
		util.Logger.Errorf("failed to find plugin directory: %s", PluginDir)
		return err
	}
	return nil
}
