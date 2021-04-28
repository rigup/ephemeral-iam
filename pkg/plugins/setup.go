package eiamplugin

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
	util "github.com/jessesomerville/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/jessesomerville/ephemeral-iam/internal/errors"
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
			return errorsutil.EiamError{
				Log: util.Logger.WithError(err),
				Msg: fmt.Sprintf("failed to create plugin directory: %s", PluginDir),
				Err: err,
			}
		}
	} else if err != nil {
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: fmt.Sprintf("failed to find plugin directory: %s", PluginDir),
			Err: err,
		}
	}
	return nil
}
