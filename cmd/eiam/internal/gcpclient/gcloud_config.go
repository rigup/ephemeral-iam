package gcpclient

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path"
	"sync"

	"gopkg.in/ini.v1"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/appconfig"
)

var (
	once         sync.Once
	gcloudConfig *ini.File
	pathToConfig string
)

func readGcloudConfigFromFile() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("Failed to get current system user: %v", err)
	}
	configDir := path.Join(usr.HomeDir, ".config", "gcloud")

	activeConfig, err := ioutil.ReadFile(path.Join(configDir, "active_config"))
	if err != nil {
		return fmt.Errorf("Unable to read %s: %v", activeConfig, err)
	}
	configName := fmt.Sprintf("config_%s", string(activeConfig))
	pathToConfig = path.Join(configDir, "configurations", configName)

	gcloudConfig, err = ini.Load(pathToConfig)
	if err != nil {
		return fmt.Errorf("Failed to parse gcloud config %s: %v", pathToConfig, err)
	}
	return nil
}

func getGcloudConfig() (configErr error) {
	once.Do(func() {
		configErr = readGcloudConfigFromFile()
	})
	return configErr
}

// ConfigureGcloudProxy configures the current gcloud configuration to use the auth proxy
func ConfigureGcloudProxy() error {
	if err := getGcloudConfig(); err != nil {
		return err
	}

	gcloudConfig.Section("proxy").Key("address").SetValue(appconfig.Config.AuthProxy.ProxyAddress)
	gcloudConfig.Section("proxy").Key("port").SetValue(appconfig.Config.AuthProxy.ProxyPort)
	gcloudConfig.Section("proxy").Key("type").SetValue("http")
	gcloudConfig.Section("core").Key("custom_ca_certs_file").SetValue(appconfig.CertFile)
	if err := gcloudConfig.SaveTo(pathToConfig); err != nil {
		return fmt.Errorf("Failed to update gcloud configuration: %v", err)
	}
	return nil
}

// UnsetGcloudProxy restores the auth proxy changes made to the gcloud config
func UnsetGcloudProxy() error {
	if err := getGcloudConfig(); err != nil {
		return err
	}

	gcloudConfig.Section("proxy").DeleteKey("address")
	gcloudConfig.Section("proxy").DeleteKey("port")
	gcloudConfig.Section("proxy").DeleteKey("type")
	gcloudConfig.Section("core").DeleteKey("custom_ca_certs_file")
	if err := gcloudConfig.SaveTo(pathToConfig); err != nil {
		return fmt.Errorf("Failed to revert gcloud configuration: %v", err)
	}
	return nil
}

// GetCurrentProject get the active project from the gcloud config
func GetCurrentProject() (string, error) {
	if err := getGcloudConfig(); err != nil {
		return "", err
	}
	return gcloudConfig.Section("core").Key("project").String(), nil
}

// GetCurrentRegion get the active region from the gcloud config
func GetCurrentRegion() (string, error) {
	if err := getGcloudConfig(); err != nil {
		return "", err
	}
	return gcloudConfig.Section("compute").Key("region").String(), nil
}

// GetCurrentZone get the active zone from the gcloud config
func GetCurrentZone() (string, error) {
	if err := getGcloudConfig(); err != nil {
		return "", err
	}
	return gcloudConfig.Section("compute").Key("zone").String(), nil
}
