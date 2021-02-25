package gcpclient

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path"
	"sync"

	"gopkg.in/ini.v1"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/appconfig"
)

type gcloudConfiguration struct {
	Project string
	Region  string
	Zone    string
}

var (
	once         sync.Once
	gcloudConfig gcloudConfiguration
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
	pathToConfig := path.Join(configDir, "configurations", configName)

	config, err := ini.Load(pathToConfig)
	if err != nil {
		return fmt.Errorf("Failed to parse gcloud config %s: %v", pathToConfig, err)
	}

	gcloudConfig.Project = config.Section("core").Key("project").String()
	gcloudConfig.Region = config.Section("compute").Key("region").String()
	gcloudConfig.Zone = config.Section("compute").Key("zone").String()
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
	if err := exec.Command("gcloud", "config", "set", "proxy/address", appconfig.Config.AuthProxy.ProxyAddress).Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "set", "proxy/port", appconfig.Config.AuthProxy.ProxyPort).Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "set", "proxy/type", "http").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "set", "core/custom_ca_certs_file", appconfig.CertFile).Run(); err != nil {
		return err
	}
	return nil
}

// UnsetGcloudProxy restores the auth proxy changes made to the gcloud config
func UnsetGcloudProxy() error {
	if err := exec.Command("gcloud", "config", "unset", "proxy/address").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "unset", "proxy/port").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "unset", "proxy/type").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "unset", "core/custom_ca_certs_file").Run(); err != nil {
		return err
	}
	return nil
}

// GetCurrentProject get the active project from the gcloud config
func GetCurrentProject() (string, error) {
	if err := getGcloudConfig(); err != nil {
		return "", err
	}
	return gcloudConfig.Project, nil
}

// GetCurrentRegion get the active region from the gcloud config
func GetCurrentRegion() (string, error) {
	if err := getGcloudConfig(); err != nil {
		return "", err
	}
	return gcloudConfig.Region, nil
}

// GetCurrentZone get the active zone from the gcloud config
func GetCurrentZone() (string, error) {
	if err := getGcloudConfig(); err != nil {
		return "", err
	}
	return gcloudConfig.Zone, nil
}
